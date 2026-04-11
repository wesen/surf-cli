package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/glazed/pkg/help"
)

func TestSurfGoPageReadCommandAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			done <- err
			return
		}
		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			done <- err
			return
		}
		resp := map[string]any{
			"type": "tool_response",
			"id":   req["id"],
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "ok"}},
			},
		}
		b, _ := json.Marshal(resp)
		_, err = conn.Write(append(b, '\n'))
		done <- err
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"page", "read", "--socket-path", sock, "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoNetworkStreamCommandStartStop(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		if _, err := reader.ReadBytes('\n'); err != nil { // stream_request
			done <- err
			return
		}
		started, _ := json.Marshal(map[string]any{"type": "stream_started", "streamId": 1})
		_, _ = conn.Write(append(started, '\n'))
		event, _ := json.Marshal(map[string]any{"type": "network_event", "method": "GET", "url": "https://example.com"})
		_, _ = conn.Write(append(event, '\n'))

		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, err = reader.ReadBytes('\n') // stream_stop
		done <- err
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"network", "stream", "--socket-path", sock, "--duration-sec", "1", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("stream command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock stream host failed: %v", err)
	}
}

func TestSurfGoDefaultOutputFormatIsYAML(t *testing.T) {
	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	tabCmd, _, err := root.Find([]string{"tab", "list"})
	if err != nil {
		t.Fatalf("failed to find tab list command: %v", err)
	}
	flag := tabCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatalf("output flag not found on tab list command")
	}
	if flag.DefValue != "yaml" {
		t.Fatalf("expected default output format yaml, got %q", flag.DefValue)
	}
}

func TestSurfGoJSCommandUsesFileInputAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	scriptPath := filepath.Join(t.TempDir(), "script.js")
	if err := os.WriteFile(scriptPath, []byte("return document.title"), 0o644); err != nil {
		t.Fatalf("write script failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			done <- err
			return
		}
		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			done <- err
			return
		}
		params := req["params"].(map[string]any)
		if params["tool"] != "js" {
			done <- fmt.Errorf("unexpected tool: %v", params["tool"])
			return
		}
		args := params["args"].(map[string]any)
		if args["code"] != "return document.title" {
			done <- fmt.Errorf("unexpected code payload: %v", args["code"])
			return
		}
		resp := map[string]any{
			"type": "tool_response",
			"id":   req["id"],
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "ok"}},
			},
		}
		b, _ := json.Marshal(resp)
		_, err = conn.Write(append(b, '\n'))
		done <- err
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"js", "--file", scriptPath, "--socket-path", sock, "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoKagiAssistantCommandUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		for i := 0; i < 4; i++ {
			conn, err := ln.Accept()
			if err != nil {
				done <- err
				return
			}

			reader := bufio.NewReader(conn)
			line, err := reader.ReadBytes('\n')
			if err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			params := req["params"].(map[string]any)

			switch i {
			case 0:
				if params["tool"] != "tab.new" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected first tool: %v", params["tool"])
					return
				}
				firstResp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":42,"url":"https://kagi.com/assistant"}`}},
					},
				}
				b, _ := json.Marshal(firstResp)
				_, err = conn.Write(append(b, '\n'))
			case 1:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected second tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, "readyState") {
					_ = conn.Close()
					done <- fmt.Errorf("expected tab-ready probe, got: %q", code)
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"href":"https://kagi.com/assistant","title":"Kagi Assistant","readyState":"complete"}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 2:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected third tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				for _, needle := range []string{`"assistant":"Quick"`, `"tags":["Temporary","photo"]`, `"createTags":true`} {
					if !strings.Contains(code, needle) {
						_ = conn.Close()
						done <- fmt.Errorf("missing %s in generated kagi-assistant code", needle)
						return
					}
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"kind":"response","href":"https://kagi.com/assistant/abc","prompt":"hello","response":"world","tagSelection":{"visibleTags":["Temporary","photo"]}}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 3:
				if params["tool"] != "tab.close" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected fourth tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				if id, ok := args["id"].(float64); !ok || int64(id) != 42 {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected close tab id: %#v", args["id"])
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":42}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			}

			_ = conn.Close()
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"kagi", "assistant", "hello", "--assistant", "Quick", "--tags", "Temporary,photo", "--create-tags", "--socket-path", sock, "--with-glaze-output", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoChatGPTTranscriptCommandUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			done <- err
			return
		}
		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			done <- err
			return
		}
		params := req["params"].(map[string]any)
		if params["tool"] != "js" {
			done <- fmt.Errorf("unexpected tool: %v", params["tool"])
			return
		}
		args := params["args"].(map[string]any)
		code, _ := args["code"].(string)
		if !strings.Contains(code, `const SURF_OPTIONS = {"withActivity":true,"activityLimit":3};`) &&
			!strings.Contains(code, `const SURF_OPTIONS = {"activityLimit":3,"withActivity":true};`) {
			done <- fmt.Errorf("missing transcript options prelude: %q", code)
			return
		}
		resp := map[string]any{
			"type": "tool_response",
			"id":   req["id"],
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": `{"href":"https://chatgpt.com/c/abc","title":"Conversation","turnCount":1,"withActivity":true,"activityLimit":3,"activityExported":1,"transcript":[{"index":0,"role":"assistant","text":"hello","activityFound":true,"activityText":"thoughts"}]}`}},
			},
		}
		b, _ := json.Marshal(resp)
		_, err = conn.Write(append(b, '\n'))
		done <- err
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"chatgpt", "transcript", "--with-activity", "--activity-limit", "3", "--socket-path", sock, "--with-glaze-output", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoClaudeCommandCreatesTabThenUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		for i := 0; i < 4; i++ {
			conn, err := ln.Accept()
			if err != nil {
				done <- err
				return
			}

			reader := bufio.NewReader(conn)
			line, err := reader.ReadBytes('\n')
			if err != nil {
				_ = conn.Close()
				done <- err
				return
			}

			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			params := req["params"].(map[string]any)

			switch i {
			case 0:
				if params["tool"] != "tab.new" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected first tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				if args["url"] != "https://claude.ai/new" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected Claude tab.new url: %v", args["url"])
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":88,"url":"https://claude.ai/new"}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 1:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected second tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, "readyState") {
					_ = conn.Close()
					done <- fmt.Errorf("expected Claude tab-ready probe, got: %q", code)
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"href":"https://claude.ai/new","title":"Claude","readyState":"complete"}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 2:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected third tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				for _, needle := range []string{`"action":"run"`, `"model":"Sonnet 4.6"`, `"prompt":"hello"`} {
					if !strings.Contains(code, needle) {
						_ = conn.Close()
						done <- fmt.Errorf("missing %s in generated Claude code", needle)
						return
					}
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"kind":"response","href":"https://claude.ai/chat/abc","conversationTitle":"Greeting","response":"Hello!","currentModel":"Sonnet 4.6"}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 3:
				if params["tool"] != "tab.close" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected fourth tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				if id, ok := args["id"].(float64); !ok || int64(id) != 88 {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected Claude close tab id: %#v", args["id"])
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":88}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			}

			_ = conn.Close()
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"claude", "ask", "hello", "--model", "Sonnet 4.6", "--socket-path", sock, "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoClaudeTranscriptCommandUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			done <- err
			return
		}
		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			done <- err
			return
		}
		params := req["params"].(map[string]any)
		if params["tool"] != "js" {
			done <- fmt.Errorf("unexpected tool: %v", params["tool"])
			return
		}
		args := params["args"].(map[string]any)
		code, _ := args["code"].(string)
		if !strings.Contains(code, "div.font-claude-response") {
			done <- fmt.Errorf("missing Claude transcript script body")
			return
		}
		resp := map[string]any{
			"type": "tool_response",
			"id":   req["id"],
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": `{"href":"https://claude.ai/chat/abc","conversationTitle":"Greeting","currentModel":"Sonnet 4.6","turnCount":2,"transcript":[{"index":0,"role":"user","text":"hello"},{"index":1,"role":"assistant","text":"world"}]}`}},
			},
		}
		b, _ := json.Marshal(resp)
		_, err = conn.Write(append(b, '\n'))
		done <- err
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"claude", "transcript", "--socket-path", sock, "--with-glaze-output", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoKagiSearchCommandCreatesTabThenUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		for i := 0; i < 4; i++ {
			conn, err := ln.Accept()
			if err != nil {
				done <- err
				return
			}

			reader := bufio.NewReader(conn)
			line, err := reader.ReadBytes('\n')
			if err != nil {
				_ = conn.Close()
				done <- err
				return
			}

			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			params := req["params"].(map[string]any)

			switch i {
			case 0:
				if params["tool"] != "tab.new" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected first tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				if args["url"] != "https://kagi.com/search?q=llm+transcript+attribution" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected tab.new url: %v", args["url"])
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":77,"url":"https://kagi.com/search?q=llm+transcript+attribution"}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 1:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected second tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, "readyState") {
					_ = conn.Close()
					done <- fmt.Errorf("expected tab-ready probe, got: %q", code)
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"href":"https://kagi.com/search?q=llm+transcript+attribution","title":"llm transcript attribution - Kagi Search","readyState":"complete"}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 2:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected third tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, `const SURF_OPTIONS = {"maxResults":3};`) {
					_ = conn.Close()
					done <- fmt.Errorf("missing kagi search options prelude: %q", code)
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"query":"llm transcript attribution","href":"https://kagi.com/search?q=llm+transcript+attribution","title":"llm transcript attribution - Kagi Search","waitedMs":500,"maxResults":3,"resultCount":1,"results":[{"index":1,"title":"Paper A","url":"https://example.com/a","snippet":"Snippet A"}]}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 3:
				if params["tool"] != "tab.close" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected fourth tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				if id, ok := args["id"].(float64); !ok || int64(id) != 77 {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected close tab id: %#v", args["id"])
					return
				}
				resp := map[string]any{
					"type": "tool_response",
					"id":   req["id"],
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":77}`}},
					},
				}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			}

			_ = conn.Close()
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"kagi", "search", "--query", "llm transcript attribution", "--max-results", "3", "--socket-path", sock, "--with-glaze-output", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoGmailListCommandCreatesTabThenUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		for i := 0; i < 4; i++ {
			conn, err := ln.Accept()
			if err != nil {
				done <- err
				return
			}

			reader := bufio.NewReader(conn)
			line, err := reader.ReadBytes('\n')
			if err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			params := req["params"].(map[string]any)

			switch i {
			case 0:
				if params["tool"] != "tab.new" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected first tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				if args["url"] != "https://mail.google.com/mail/u/0/#inbox" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected tab.new url: %v", args["url"])
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":88,"url":"https://mail.google.com/mail/u/0/#inbox"}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 1:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected second tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, "readyState") {
					_ = conn.Close()
					done <- fmt.Errorf("expected tab-ready probe, got: %q", code)
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"href":"https://mail.google.com/mail/u/0/#inbox","title":"Inbox - Gmail","readyState":"complete"}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 2:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected third tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, `"mailbox":"inbox"`) {
					_ = conn.Close()
					done <- fmt.Errorf("missing gmail list inbox options: %q", code)
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"href":"https://mail.google.com/mail/u/0/#inbox","title":"Inbox - Gmail","mailbox":"inbox","waitedMs":300,"resultCount":1,"threads":[{"index":1,"threadId":"#thread-f:1","legacyThreadId":"abc","participant":"GitHub","subject":"Security alert","snippet":"Snippet","timestamp":"7:31 AM","unread":true,"starred":false,"hasAttachment":false}]}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 3:
				if params["tool"] != "tab.close" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected fourth tool: %v", params["tool"])
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":88}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			}

			_ = conn.Close()
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"gmail", "list", "--inbox", "--max-results", "3", "--socket-path", sock, "--with-glaze-output", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}

func TestSurfGoGmailSearchCommandCreatesTabThenUsesJSAgainstMockHost(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		for i := 0; i < 4; i++ {
			conn, err := ln.Accept()
			if err != nil {
				done <- err
				return
			}

			reader := bufio.NewReader(conn)
			line, err := reader.ReadBytes('\n')
			if err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				_ = conn.Close()
				done <- err
				return
			}
			params := req["params"].(map[string]any)

			switch i {
			case 0:
				if params["tool"] != "tab.new" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected first tool: %v", params["tool"])
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":89,"url":"https://mail.google.com/mail/u/0/#inbox"}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 1:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected second tool: %v", params["tool"])
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"href":"https://mail.google.com/mail/u/0/#inbox","title":"Inbox - Gmail","readyState":"complete"}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 2:
				if params["tool"] != "js" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected third tool: %v", params["tool"])
					return
				}
				args := params["args"].(map[string]any)
				code, _ := args["code"].(string)
				if !strings.Contains(code, `"query":"from:boss"`) {
					_ = conn.Close()
					done <- fmt.Errorf("missing gmail search query options: %q", code)
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"href":"https://mail.google.com/mail/u/0/#search/from%3Aboss","title":"Search results - Gmail","query":"from:boss","waitedMs":350,"resultCount":1,"threads":[{"index":1,"threadId":"#thread-f:2","legacyThreadId":"def","participant":"Boss","subject":"Follow-up","snippet":"Please send the draft","timestamp":"9:10 AM","unread":false,"starred":false,"hasAttachment":true}]}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			case 3:
				if params["tool"] != "tab.close" {
					_ = conn.Close()
					done <- fmt.Errorf("unexpected fourth tool: %v", params["tool"])
					return
				}
				resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": `{"success":true,"tabId":89}`}}}}
				b, _ := json.Marshal(resp)
				_, err = conn.Write(append(b, '\n'))
			}

			_ = conn.Close()
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	root, err := newRootCommand(help.NewHelpSystem())
	if err != nil {
		t.Fatalf("failed to build root: %v", err)
	}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"gmail", "search", "from:boss", "--max-results", "3", "--socket-path", sock, "--with-glaze-output", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("mock host failed: %v", err)
	}
}
