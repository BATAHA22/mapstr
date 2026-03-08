package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mapstr/mapstr/internal/config"
	"github.com/mapstr/mapstr/internal/graph"
	"github.com/mapstr/mapstr/internal/output"
	"github.com/mapstr/mapstr/internal/parser"
)

// MCP JSON-RPC types
type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *mcpError `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(os.Stderr, "Mapstr MCP server started (stdio)")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal(line, &req); err != nil {
			writeResponse(mcpResponse{
				JSONRPC: "2.0",
				Error:   &mcpError{Code: -32700, Message: "parse error"},
			})
			continue
		}

		resp := handleMCPRequest(req)
		writeResponse(resp)
	}

	return scanner.Err()
}

func handleMCPRequest(req mcpRequest) mcpResponse {
	switch req.Method {
	case "initialize":
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "mapstr",
					"version": Version,
				},
			},
		}

	case "tools/list":
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": []map[string]any{
					{
						"name":        "analyze",
						"description": "Analyze a codebase and return structured context",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"path": map[string]any{
									"type":        "string",
									"description": "Path to the project directory",
								},
								"format": map[string]any{
									"type":        "string",
									"description": "Output format: md, json, mermaid",
									"default":     "json",
								},
							},
							"required": []string{"path"},
						},
					},
				},
			},
		}

	case "tools/call":
		return handleToolCall(req)

	case "notifications/initialized":
		// Client acknowledgement, no response needed
		return mcpResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}

	default:
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)},
		}
	}
}

func handleToolCall(req mcpRequest) mcpResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32602, Message: "invalid params"},
		}
	}

	if params.Name != "analyze" {
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", params.Name)},
		}
	}

	var toolArgs struct {
		Path   string `json:"path"`
		Format string `json:"format"`
	}
	if err := json.Unmarshal(params.Arguments, &toolArgs); err != nil {
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32602, Message: "invalid tool arguments"},
		}
	}

	if toolArgs.Format == "" {
		toolArgs.Format = "json"
	}

	absPath, err := filepath.Abs(toolArgs.Path)
	if err != nil {
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32000, Message: err.Error()},
		}
	}

	cfg := config.DefaultConfig()
	cfg.AI.NoAI = true // MCP mode uses structural analysis only

	nodes, err := parser.ParseProject(absPath, cfg)
	if err != nil {
		return mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32000, Message: err.Error()},
		}
	}

	g := graph.Build(nodes)
	projectName := filepath.Base(absPath)

	var content string
	switch toolArgs.Format {
	case "md", "markdown":
		content = output.GenerateMarkdown(projectName, nodes, g, "")
	case "mermaid", "mmd":
		content = output.GenerateMermaid(g)
	default:
		jsonContent, jsonErr := output.GenerateJSON(projectName, nodes, g, "")
		if jsonErr != nil {
			return mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: -32000, Message: jsonErr.Error()},
			}
		}
		content = jsonContent
	}

	return mcpResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": content,
				},
			},
		},
	}
}

func writeResponse(resp mcpResponse) {
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}
