package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"runtime"
	"time"
)

type Client struct {
	SocketPath string
	Timeout    time.Duration
}

func NewClient(socketPath string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{SocketPath: socketPath, Timeout: timeout}
}

func (c *Client) Send(ctx context.Context, req map[string]any) (map[string]any, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("windows named pipe transport is not implemented yet")
	}

	dialer := net.Dialer{Timeout: c.Timeout}
	conn, err := dialer.DialContext(ctx, "unix", c.SocketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	payload = append(payload, '\n')
	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var resp map[string]any
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Stream(
	ctx context.Context,
	streamType string,
	options map[string]any,
	tabID *int64,
	onMessage func(map[string]any) error,
) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("windows named pipe transport is not implemented yet")
	}
	if options == nil {
		options = map[string]any{}
	}

	dialer := net.Dialer{Timeout: c.Timeout}
	conn, err := dialer.DialContext(ctx, "unix", c.SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := map[string]any{
		"type":       "stream_request",
		"streamType": streamType,
		"options":    options,
		"id":         fmt.Sprintf("go-stream-%d", time.Now().UnixNano()),
	}
	if tabID != nil {
		req["tabId"] = *tabID
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if _, err := conn.Write(payload); err != nil {
		return err
	}

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			stop := []byte("{\"type\":\"stream_stop\"}\n")
			_, _ = conn.Write(stop)
			return nil
		default:
		}

		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		var msg map[string]any
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg["type"] == "stream_started" {
			continue
		}
		if msg["type"] == "extension_disconnected" {
			return fmt.Errorf("%v", msg["message"])
		}
		if onMessage != nil {
			if err := onMessage(msg); err != nil {
				return err
			}
		}
	}
}
