package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nicobailon/surf-cli/gohost/internal/cli/commands"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
	"github.com/spf13/cobra"
)

type chatGPTCommandSettings struct {
	Model      string
	File       string
	WithPage   bool
	ListModels bool
	TimeoutSec int
	SocketPath string
	TabID      int64
	WindowID   int64
	JSON       bool
}

func newChatGPTCommand() *cobra.Command {
	settings := &chatGPTCommandSettings{
		TimeoutSec: 2700,
		SocketPath: config.CurrentSocketPath(),
		TabID:      -1,
		WindowID:   -1,
	}

	cmd := &cobra.Command{
		Use:   "chatgpt [query]",
		Short: "Send a prompt to ChatGPT using your browser session",
		Long: "Uses the Go native host ChatGPT provider through the local surf socket. " +
			"Requires an active ChatGPT browser login.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = strings.TrimSpace(args[0])
			}
			if query == "" && !settings.ListModels {
				return fmt.Errorf("query required unless --list-models is set")
			}

			toolArgs := map[string]any{}
			if query != "" {
				toolArgs["query"] = query
			}
			if strings.TrimSpace(settings.Model) != "" {
				toolArgs["model"] = strings.TrimSpace(settings.Model)
			}
			if strings.TrimSpace(settings.File) != "" {
				toolArgs["file"] = strings.TrimSpace(settings.File)
			}
			if settings.WithPage {
				toolArgs["with-page"] = true
			}
			if settings.ListModels {
				toolArgs["list-models"] = true
			}
			if settings.TimeoutSec > 0 {
				toolArgs["timeout"] = settings.TimeoutSec
			}

			client := transport.NewClient(settings.SocketPath, time.Duration(settings.TimeoutSec+30)*time.Second)
			var tabID *int64
			if settings.TabID >= 0 {
				tabID = &settings.TabID
			}
			var windowID *int64
			if settings.WindowID >= 0 {
				windowID = &settings.WindowID
			}

			resp, err := commands.ExecuteTool(context.Background(), client, "chatgpt", toolArgs, tabID, windowID)
			if err != nil {
				return err
			}
			data, err := decodeToolPayload(resp)
			if err != nil {
				return err
			}

			if settings.JSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}
			return renderChatGPTResult(data)
		},
	}

	cmd.Flags().StringVar(&settings.Model, "model", "", "ChatGPT model to select")
	cmd.Flags().StringVar(&settings.File, "file", "", "File to attach before sending the prompt")
	cmd.Flags().BoolVar(&settings.WithPage, "with-page", false, "Include current page content in the prompt")
	cmd.Flags().BoolVar(&settings.ListModels, "list-models", false, "List available ChatGPT models without sending a prompt")
	cmd.Flags().IntVar(&settings.TimeoutSec, "timeout", settings.TimeoutSec, "Request timeout in seconds")
	cmd.Flags().StringVar(&settings.SocketPath, "socket-path", settings.SocketPath, "Host socket path")
	cmd.Flags().Int64Var(&settings.TabID, "tab-id", settings.TabID, "Optional tab id override for page-context reads")
	cmd.Flags().Int64Var(&settings.WindowID, "window-id", settings.WindowID, "Optional window id override")
	cmd.Flags().BoolVar(&settings.JSON, "json", false, "Emit raw JSON result")
	return cmd
}

func decodeToolPayload(resp map[string]any) (map[string]any, error) {
	if resp == nil {
		return nil, fmt.Errorf("empty response from host")
	}
	if errText := extractToolError(resp); errText != "" {
		return nil, fmt.Errorf("%s", errText)
	}
	result, _ := resp["result"].(map[string]any)
	content, _ := result["content"].([]any)
	if len(content) == 0 {
		return map[string]any{}, nil
	}
	block, _ := content[0].(map[string]any)
	text := strings.TrimSpace(stringValue(block["text"]))
	if text == "" {
		return map[string]any{}, nil
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err == nil {
		return data, nil
	}
	return map[string]any{"response": text}, nil
}

func extractToolError(resp map[string]any) string {
	rawErr, ok := resp["error"]
	if !ok || rawErr == nil {
		return ""
	}
	if s, ok := rawErr.(string); ok {
		return s
	}
	em, ok := rawErr.(map[string]any)
	if !ok {
		return fmt.Sprintf("%v", rawErr)
	}
	content, ok := em["content"].([]any)
	if !ok || len(content) == 0 {
		b, _ := json.Marshal(em)
		return string(b)
	}
	block, ok := content[0].(map[string]any)
	if !ok {
		return "unknown error"
	}
	return strings.TrimSpace(stringValue(block["text"]))
}

func renderChatGPTResult(data map[string]any) error {
	if models, ok := data["models"].([]any); ok {
		selected := strings.TrimSpace(stringValue(data["selected"]))
		if len(models) == 0 {
			fmt.Println("No ChatGPT models found")
			return nil
		}
		for _, model := range models {
			name := strings.TrimSpace(stringValue(model))
			if name == "" {
				continue
			}
			if selected != "" && selected == name {
				fmt.Printf("* %s (selected)\n", name)
			} else {
				fmt.Printf("* %s\n", name)
			}
		}
		return nil
	}

	if response := strings.TrimSpace(stringValue(data["response"])); response != "" {
		fmt.Println(response)
		return nil
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}
