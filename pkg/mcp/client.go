package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPSourceConfig configures an MCP tool source.
type MCPSourceConfig struct {
	// Type: "stdio" or "http"
	Type string `yaml:"type"`
	// For stdio: command to run (e.g. "npx", "node")
	Command string `yaml:"command,omitempty"`
	// For stdio: args (e.g. ["-y", "mcp-server-filesystem"])
	Args []string `yaml:"args,omitempty"`
	// For http: base URL (e.g. "https://mcp.example.com")
	URL string `yaml:"url,omitempty"`
	// Optional prefix for tool names (e.g. "mcp_filesystem_" to avoid clashes)
	ToolPrefix string `yaml:"tool_prefix,omitempty"`
}

// MCPClient wraps an MCP client and provides tool listing/calling.
type MCPClient struct {
	client *mcpclient.Client
	prefix string
}

// NewMCPClient creates an MCP client from config.
func NewMCPClient(ctx context.Context, cfg MCPSourceConfig) (*MCPClient, error) {
	var c *mcpclient.Client
	var err error

	switch cfg.Type {
	case "stdio":
		if cfg.Command == "" {
			return nil, fmt.Errorf("mcp stdio: command is required")
		}
		c, err = mcpclient.NewStdioMCPClient(cfg.Command, nil, cfg.Args...)
		if err != nil {
			return nil, fmt.Errorf("mcp stdio client: %w", err)
		}
	case "http":
		if cfg.URL == "" {
			return nil, fmt.Errorf("mcp http: url is required")
		}
		c, err = mcpclient.NewStreamableHttpClient(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("mcp http client: %w", err)
		}
	default:
		return nil, fmt.Errorf("mcp: unsupported type %q (use stdio or http)", cfg.Type)
	}

	// Initialize MCP session
	initReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo:      mcp.Implementation{Name: "agentruntime", Version: "1.0"},
		},
	}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		c.Close()
		return nil, fmt.Errorf("mcp initialize: %w", err)
	}

	return &MCPClient{client: c, prefix: cfg.ToolPrefix}, nil
}

// ListTools returns tools from the MCP server.
func (m *MCPClient) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	res, err := m.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	return res.Tools, nil
}

// CallTool invokes an MCP tool.
func (m *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Strip prefix when calling
	callName := name
	if m.prefix != "" && len(name) > len(m.prefix) && name[:len(m.prefix)] == m.prefix {
		callName = name[len(m.prefix):]
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      callName,
			Arguments: arguments,
		},
	}
	return m.client.CallTool(ctx, req)
}

// PrefixedName returns the tool name with prefix for registration.
func (m *MCPClient) PrefixedName(name string) string {
	if m.prefix == "" {
		return name
	}
	return m.prefix + name
}

// Close closes the MCP client connection.
func (m *MCPClient) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

// ToolToInputSchema converts MCP tool input schema to JSON.
func ToolToInputSchema(t mcp.Tool) ([]byte, error) {
	if len(t.RawInputSchema) > 0 {
		return t.RawInputSchema, nil
	}
	if t.InputSchema.Type != "" || len(t.InputSchema.Properties) > 0 {
		return json.Marshal(t.InputSchema)
	}
	return json.Marshal(map[string]interface{}{"type": "object", "properties": map[string]interface{}{}})
}
