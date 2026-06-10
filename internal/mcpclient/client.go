// Package mcpclient 把官方 go-sdk 的 MCP 客户端封装成一个面向
// "传参数调工具" 的薄接口，CLI 层只需要关心工具名和参数。
package mcpclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yu/openluckin/internal/config"
)

// Client 持有到瑞幸 MCP 服务的一条会话。
type Client struct {
	cfg     config.Config
	session *mcp.ClientSession
}

// ToolInfo 是对外暴露的工具元信息（用于 tools 子命令）。
type ToolInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	InputSchema  any    `json:"inputSchema,omitempty"`
	OutputSchema any    `json:"outputSchema,omitempty"`
}

// New 创建一个尚未连接的客户端。
func New(cfg config.Config) *Client {
	return &Client{cfg: cfg}
}

// bearerTransport 给每个请求加上 Authorization 头。
type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

// Connect 建立 MCP 会话（streamable HTTP）。
func (c *Client) Connect(ctx context.Context) error {
	if err := c.cfg.Validate(); err != nil {
		return err
	}
	httpClient := &http.Client{
		Timeout: c.cfg.Timeout,
		Transport: &bearerTransport{
			token: c.cfg.Token,
			base:  http.DefaultTransport,
		},
	}
	transport := &mcp.StreamableClientTransport{
		Endpoint:   c.cfg.Endpoint,
		HTTPClient: httpClient,
	}
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "openluckin",
		Version: "0.1.0",
	}, nil)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("连接 MCP 服务失败: %w", err)
	}
	c.session = session
	return nil
}

// ListTools 返回服务端暴露的所有工具。
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error) {
	if c.session == nil {
		return nil, fmt.Errorf("尚未连接，请先调用 Connect")
	}
	res, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("获取工具列表失败: %w", err)
	}
	tools := make([]ToolInfo, 0, len(res.Tools))
	for _, t := range res.Tools {
		tools = append(tools, ToolInfo{
			Name:         t.Name,
			Description:  t.Description,
			InputSchema:  t.InputSchema,
			OutputSchema: t.OutputSchema,
		})
	}
	return tools, nil
}

// CallTool 以 map 形式传参调用指定工具，返回文本结果。
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	if c.session == nil {
		return "", fmt.Errorf("尚未连接，请先调用 Connect")
	}
	res, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", fmt.Errorf("调用工具 %s 失败: %w", name, err)
	}

	var sb strings.Builder
	for _, content := range res.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	if res.IsError {
		return "", fmt.Errorf("工具 %s 返回错误: %s", name, sb.String())
	}
	return sb.String(), nil
}

// Close 关闭会话。
func (c *Client) Close() error {
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}
