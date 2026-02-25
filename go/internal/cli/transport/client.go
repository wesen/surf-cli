package transport

import (
	"bufio"
	"context"
	"encoding/json"
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
