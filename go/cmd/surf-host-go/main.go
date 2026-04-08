package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
	"github.com/nicobailon/surf-cli/gohost/internal/host/nativeio"
	"github.com/nicobailon/surf-cli/gohost/internal/host/pending"
	"github.com/nicobailon/surf-cli/gohost/internal/host/providers"
	"github.com/nicobailon/surf-cli/gohost/internal/host/router"
	"github.com/nicobailon/surf-cli/gohost/internal/host/socketbridge"
)

var errNativeDisconnected = errors.New("native stdin disconnected")

type hostRuntime struct {
	log      *log.Logger
	decoder  *nativeio.Decoder
	sessions *socketbridge.SessionManager
	pending  *pending.Store
	streams  *router.StreamRegistry
	ids      *pending.IDAllocator

	nativeWriteMu sync.Mutex

	providerPendingMu sync.Mutex
	providerPending   map[int64]chan map[string]any

	runChatGPTTool func(ctx context.Context, args map[string]any, tabID *int64, logf func(string, ...any)) (map[string]any, error)
}

func main() {
	logger, closer := newLogger()
	if closer != nil {
		defer closer.Close()
	}

	logger.Printf("Host starting...")
	if err := run(logger); err != nil {
		logger.Printf("Host failed: %v", err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	endpoint := config.CurrentSocketPath()
	listener, err := socketbridge.Listen(endpoint)
	if err != nil {
		return err
	}
	defer listener.Close()
	defer listener.Cleanup()

	h := &hostRuntime{
		log:             logger,
		decoder:         nativeio.NewDecoder(os.Stdin, 0),
		sessions:        socketbridge.NewSessionManager(),
		pending:         pending.NewStore(),
		streams:         router.NewStreamRegistry(),
		ids:             pending.NewIDAllocator(0),
		providerPending: map[int64]chan map[string]any{},
	}
	h.runChatGPTTool = h.defaultChatGPTToolRunner

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	acceptErrCh := make(chan error, 1)
	go func() {
		acceptErrCh <- h.acceptLoop(ctx, listener)
	}()

	readErrCh := make(chan error, 1)
	go func() {
		readErrCh <- h.readNativeLoop(ctx)
	}()

	if err := h.writeNative(map[string]any{
		"type":       "HOST_READY",
		"runtime":    "go-host",
		"socketPath": listener.Endpoint(),
	}); err != nil {
		return err
	}
	logger.Printf("Host initialization complete, waiting for connections...")
	logger.Printf("Socket server listening on %s", listener.Endpoint())
	logger.Printf("Sent HOST_READY to extension")

	for {
		select {
		case sig := <-sigCh:
			logger.Printf("%s received", sig.String())
			cancel()
			h.sessions.CloseAll()
			return nil
		case err := <-acceptErrCh:
			if err == nil || errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) {
				acceptErrCh = nil
				continue
			}
			return err
		case err := <-readErrCh:
			if errors.Is(err, errNativeDisconnected) {
				logger.Printf("stdin ended (extension disconnected), notifying clients")
				h.sessions.NotifyExtensionDisconnected("Surf extension was reloaded. Restart your command.")
				cancel()
				return nil
			}
			if err != nil {
				return err
			}
			return nil
		case <-ctx.Done():
			h.sessions.CloseAll()
			return nil
		}
	}
}

func (h *hostRuntime) acceptLoop(ctx context.Context, listener socketbridge.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				h.log.Printf("temporary accept error: %v", err)
				continue
			}
			return err
		}

		session := h.sessions.Add(conn)
		h.log.Printf("CLI client connected")
		go h.handleSession(ctx, session)
	}
}

func (h *hostRuntime) handleSession(ctx context.Context, session *socketbridge.Session) {
	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer h.disconnectSession(session)

	reader := bufio.NewReader(session.Conn())
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			h.handleSessionLine(sessionCtx, session, line)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			h.log.Printf("CLI socket read error: %v", err)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (h *hostRuntime) handleSessionLine(ctx context.Context, session *socketbridge.Session, line []byte) {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return
	}

	var msg map[string]any
	if err := json.Unmarshal(trimmed, &msg); err != nil {
		h.writeClientError(session, "Invalid request")
		return
	}

	msgType := asString(msg["type"])
	switch msgType {
	case "tool_request":
		req, err := router.ParseToolRequest(msg)
		if err != nil {
			h.sendToolError(session, msg["id"], err.Error())
			return
		}
		if strings.TrimSpace(req.Params.Tool) == "chatgpt" {
			go func() {
				result, err := h.runChatGPTTool(ctx, req.Params.Args, req.TabID, h.log.Printf)
				if err != nil {
					h.sendToolError(session, requestOriginalID(req), err.Error())
					return
				}
				h.sendToolResult(session, requestOriginalID(req), result)
			}()
			return
		}

		extensionMsg, err := router.MapToolToMessage(req)
		if err != nil {
			h.sendToolError(session, requestOriginalID(req), err.Error())
			return
		}

		if asString(extensionMsg["type"]) == "UNSUPPORTED_ACTION" {
			h.sendToolError(session, requestOriginalID(req), asString(extensionMsg["message"]))
			return
		}
		if asString(extensionMsg["type"]) == "LOCAL_WAIT" {
			seconds, ok := asInt64(extensionMsg["seconds"])
			if !ok || seconds < 1 {
				seconds = 1
			}
			if seconds > 30 {
				seconds = 30
			}
			orig := requestOriginalID(req)
			time.AfterFunc(time.Duration(seconds)*time.Second, func() {
				h.sendToolResult(session, orig, map[string]any{"success": true})
			})
			return
		}

		hostID := h.ids.Next()
		p := pending.Request{
			Session: session,
			Kind:    pending.KindToolRequest,
		}
		if req.HasOriginalID {
			p.OriginalID = req.OriginalID
			p.HasOriginalID = true
		}
		h.pending.Put(hostID, p)
		extensionMsg["id"] = hostID

		if err := h.writeNative(extensionMsg); err != nil {
			_, _ = h.pending.Pop(hostID)
			h.sendToolError(session, requestOriginalID(req), "Native host unavailable")
		}
		return
	case "stream_request":
		req, err := router.ParseStreamRequest(msg)
		if err != nil {
			h.writeClientError(session, err.Error())
			return
		}

		streamID := h.ids.Next()
		h.streams.Set(streamID, session)
		forward := map[string]any{
			"type":     req.StreamType,
			"streamId": streamID,
			"options":  req.Options,
		}
		if tabID, ok := asInt64(msg["tabId"]); ok {
			forward["tabId"] = tabID
		}
		if err := h.writeNative(forward); err != nil {
			h.streams.Delete(streamID)
			h.writeClientError(session, "Native host unavailable")
			return
		}
		if err := session.WriteJSONLine(map[string]any{"type": "stream_started", "streamId": streamID}); err != nil {
			h.disconnectSession(session)
		}
		return
	case "stream_stop":
		if err := router.ParseStreamStop(msg); err != nil {
			h.writeClientError(session, err.Error())
			return
		}
		h.stopStreamsForSession(session)
		return
	}

	hostID := h.ids.Next()
	req := pending.Request{
		Session: session,
		Kind:    toRequestKind(msgType),
	}
	if id, ok := msg["id"]; ok {
		req.OriginalID = id
		req.HasOriginalID = true
	}
	h.pending.Put(hostID, req)

	msg["id"] = hostID
	if err := h.writeNative(msg); err != nil {
		_, _ = h.pending.Pop(hostID)
		h.writeClientError(session, "Native host unavailable")
	}
}

func (h *hostRuntime) disconnectSession(session *socketbridge.Session) {
	h.sessions.Remove(session)
	_ = session.Close()
	_ = h.pending.DeleteForSession(session)
	h.stopStreamsForSession(session)
	h.log.Printf("CLI client disconnected")
}

func (h *hostRuntime) stopStreamsForSession(session *socketbridge.Session) {
	for _, streamID := range h.streams.DeleteBySession(session) {
		if err := h.writeNative(map[string]any{"type": "STREAM_STOP", "streamId": streamID}); err != nil {
			h.log.Printf("failed to send STREAM_STOP for stream %d: %v", streamID, err)
		}
	}
}

func (h *hostRuntime) readNativeLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var msg map[string]any
		err := h.decoder.ReadJSON(&msg)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return errNativeDisconnected
			}
			if errors.Is(err, nativeio.ErrInvalidJSON) || errors.Is(err, nativeio.ErrFrameTooLarge) {
				h.log.Printf("discarding malformed native frame: %v", err)
				continue
			}
			return err
		}
		h.handleNativeMessage(msg)
	}
}

func (h *hostRuntime) handleNativeMessage(msg map[string]any) {
	msgType := asString(msg["type"])
	if msgType == "STREAM_EVENT" {
		streamID, ok := asInt64(msg["streamId"])
		if !ok {
			return
		}
		session, ok := h.streams.Get(streamID)
		if !ok {
			return
		}
		if err := session.WriteJSONLine(msg["event"]); err != nil {
			h.disconnectSession(session)
		}
		return
	}

	if msgType == "STREAM_ERROR" {
		streamID, ok := asInt64(msg["streamId"])
		if !ok {
			return
		}
		session, ok := h.streams.Get(streamID)
		if !ok {
			return
		}
		h.streams.Delete(streamID)
		if err := session.WriteJSONLine(map[string]any{"error": msg["error"]}); err != nil {
			h.disconnectSession(session)
		}
		return
	}

	id, ok := asInt64(msg["id"])
	if !ok {
		return
	}
	if ch, ok := h.popProviderPending(id); ok {
		select {
		case ch <- msg:
		default:
		}
		return
	}
	req, ok := h.pending.Pop(id)
	if !ok {
		return
	}

	if req.Kind == pending.KindToolRequest {
		originalID := any(nil)
		if req.HasOriginalID {
			originalID = req.OriginalID
		}
		payload := stripInternalID(msg)
		if errText, pure := pureExtensionError(payload); pure {
			h.sendToolError(req.Session, originalID, errText)
			return
		}
		h.sendToolResult(req.Session, originalID, payload)
		return
	}

	if req.HasOriginalID {
		msg["id"] = req.OriginalID
	} else {
		delete(msg, "id")
	}

	if err := req.Session.WriteJSONLine(msg); err != nil {
		h.disconnectSession(req.Session)
	}
}

func (h *hostRuntime) writeNative(msg any) error {
	h.nativeWriteMu.Lock()
	defer h.nativeWriteMu.Unlock()
	return nativeio.WriteJSON(os.Stdout, msg)
}

func (h *hostRuntime) requestNativeForProvider(ctx context.Context, msg map[string]any, timeout time.Duration) (map[string]any, error) {
	hostID := h.ids.Next()
	replyCh := make(chan map[string]any, 1)
	h.putProviderPending(hostID, replyCh)

	clone := cloneMap(msg)
	clone["id"] = hostID
	if err := h.writeNative(clone); err != nil {
		h.popProviderPending(hostID)
		return nil, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		h.popProviderPending(hostID)
		return nil, ctx.Err()
	case <-timer.C:
		h.popProviderPending(hostID)
		return nil, fmt.Errorf("timeout waiting for extension: %s", asString(msg["type"]))
	case resp := <-replyCh:
		return stripInternalID(resp), nil
	}
}

func (h *hostRuntime) defaultChatGPTToolRunner(ctx context.Context, args map[string]any, tabID *int64, logf func(string, ...any)) (map[string]any, error) {
	caller := providerNativeCaller{runtime: h}
	return providers.HandleChatGPTTool(ctx, caller, args, tabID, logf)
}

type providerNativeCaller struct {
	runtime *hostRuntime
}

func (c providerNativeCaller) Request(ctx context.Context, msg map[string]any, timeout time.Duration) (map[string]any, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return c.runtime.requestNativeForProvider(ctx, msg, timeout)
}

func (h *hostRuntime) putProviderPending(id int64, ch chan map[string]any) {
	h.providerPendingMu.Lock()
	h.providerPending[id] = ch
	h.providerPendingMu.Unlock()
}

func (h *hostRuntime) popProviderPending(id int64) (chan map[string]any, bool) {
	h.providerPendingMu.Lock()
	defer h.providerPendingMu.Unlock()
	ch, ok := h.providerPending[id]
	if ok {
		delete(h.providerPending, id)
	}
	return ch, ok
}

func (h *hostRuntime) writeClientError(session *socketbridge.Session, message string) {
	if err := session.WriteJSONLine(map[string]any{"error": message}); err != nil {
		h.log.Printf("failed to write client error: %v", err)
	}
}

func (h *hostRuntime) sendToolError(session *socketbridge.Session, id any, message string) {
	resp := map[string]any{
		"type": "tool_response",
		"id":   id,
		"error": map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": message},
			},
		},
	}
	if err := session.WriteJSONLine(resp); err != nil {
		h.log.Printf("failed to write tool error: %v", err)
	}
}

func (h *hostRuntime) sendToolResult(session *socketbridge.Session, id any, result any) {
	resp := map[string]any{
		"type": "tool_response",
		"id":   id,
		"result": map[string]any{
			"content": formatToolContent(result),
		},
	}
	if err := session.WriteJSONLine(resp); err != nil {
		h.log.Printf("failed to write tool result: %v", err)
	}
}

func requestOriginalID(req router.ToolRequest) any {
	if req.HasOriginalID {
		return req.OriginalID
	}
	return nil
}

func toRequestKind(msgType string) pending.RequestKind {
	switch msgType {
	case "tool_request":
		return pending.KindToolRequest
	case "stream_request":
		return pending.KindStream
	default:
		return pending.KindUnknown
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case json.Number:
		x, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return x, true
	default:
		return 0, false
	}
}

func stripInternalID(msg map[string]any) map[string]any {
	out := make(map[string]any, len(msg))
	for k, v := range msg {
		if k == "id" {
			continue
		}
		out[k] = v
	}
	return out
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func pureExtensionError(msg map[string]any) (string, bool) {
	rawErr, ok := msg["error"]
	if !ok {
		return "", false
	}
	for k, v := range msg {
		if k == "type" || k == "error" {
			continue
		}
		if v != nil {
			return "", false
		}
	}
	switch e := rawErr.(type) {
	case string:
		return e, true
	default:
		b, err := json.Marshal(e)
		if err != nil {
			return fmt.Sprintf("%v", e), true
		}
		return string(b), true
	}
}

func formatToolContent(result any) []map[string]any {
	text := func(s string) []map[string]any {
		return []map[string]any{{"type": "text", "text": s}}
	}
	if result == nil {
		return text("OK")
	}

	if m, ok := result.(map[string]any); ok {
		for _, key := range []string{"output", "message", "text", "pageContent"} {
			if s, ok := m[key].(string); ok && s != "" {
				return text(s)
			}
		}
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return text(fmt.Sprintf("%v", result))
	}
	return text(string(b))
}

func newLogger() (*log.Logger, io.Closer) {
	path := os.Getenv("SURF_HOST_LOG")
	if path == "" {
		path = "/tmp/surf-host-go.log"
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return log.New(os.Stderr, "", log.LstdFlags|log.LUTC|log.Lmicroseconds), nil
	}
	return log.New(f, "", log.LstdFlags|log.LUTC|log.Lmicroseconds), f
}
