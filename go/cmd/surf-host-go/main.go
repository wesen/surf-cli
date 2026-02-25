package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
	"github.com/nicobailon/surf-cli/gohost/internal/host/nativeio"
	"github.com/nicobailon/surf-cli/gohost/internal/host/pending"
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
		log:      logger,
		decoder:  nativeio.NewDecoder(os.Stdin, 0),
		sessions: socketbridge.NewSessionManager(),
		pending:  pending.NewStore(),
		streams:  router.NewStreamRegistry(),
		ids:      pending.NewIDAllocator(0),
	}

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

	if err := h.writeNative(map[string]any{"type": "HOST_READY"}); err != nil {
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
	defer h.disconnectSession(session)

	reader := bufio.NewReader(session.Conn())
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			h.handleSessionLine(session, line)
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

func (h *hostRuntime) handleSessionLine(session *socketbridge.Session, line []byte) {
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
	if msgType == "stream_stop" {
		h.stopStreamsForSession(session)
		return
	}

	hostID := h.ids.Next()
	req := pending.Request{
		Session: session,
		Kind:    toRequestKind(msgType),
	}
	if id, ok := asInt64(msg["id"]); ok {
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
	req, ok := h.pending.Pop(id)
	if !ok {
		return
	}

	if req.Kind == pending.KindStream && msgType == "STREAM_STARTED" {
		if streamID, ok := asInt64(msg["streamId"]); ok {
			h.streams.Set(streamID, req.Session)
		}
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

func (h *hostRuntime) writeClientError(session *socketbridge.Session, message string) {
	if err := session.WriteJSONLine(map[string]any{"error": message}); err != nil {
		h.log.Printf("failed to write client error: %v", err)
	}
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
