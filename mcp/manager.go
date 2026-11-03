package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type Config struct {
	McpServers map[string]ServerConfig `json:"mcpServers"`
}

type Server struct {
	Name    string
	Session *mcp.ClientSession
	Tools   []*mcp.Tool
}

type Manager struct {
	servers []*Server
}

type DiscoveredTool struct {
	Server  string
	Session *mcp.ClientSession
	Tool    *mcp.Tool
}

// NewManager 读取 mcp.json 并连接所有服务
func NewManager(ctx context.Context, configPath string) (*Manager, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manager{}, nil // 没有配置就返回空管理器
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 mcp.json 失败: %w", err)
	}

	mgr := &Manager{}
	for name, sc := range cfg.McpServers {
		if sc.Command == "" {
			continue
		}

		server, err := connectStdio(ctx, name, sc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[mcp] 连接 %s 失败（已跳过）: %v\n", name, err)
			continue
		}
		mgr.servers = append(mgr.servers, server)
		fmt.Fprintf(os.Stderr, "[mcp] %s 已连接，发现 %d 个工具\n", name, len(server.Tools))
	}
	return mgr, nil
}

func connectStdio(ctx context.Context, name string, sc ServerConfig) (*Server, error) {
	cmd := exec.Command(sc.Command, sc.Args...)
	if len(sc.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range sc.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "mini-time-agent", Version: "0.1.0"},
		nil,
	)
	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("获取工具列表失败: %w", err)
	}

	return &Server{
		Name:    name,
		Session: session,
		Tools:   result.Tools,
	}, nil
}

func (m *Manager) AllTools() []DiscoveredTool {
	var out []DiscoveredTool
	for _, s := range m.servers {
		for _, t := range s.Tools {
			out = append(out, DiscoveredTool{
				Server:  s.Name,
				Session: s.Session,
				Tool:    t,
			})
		}
	}
	return out
}

func (m *Manager) Close() {
	for _, s := range m.servers {
		if s.Session != nil {
			_ = s.Session.Close()
		}
	}
}

// SanitizeName 防止不同 MCP 服务的工具重名，也规避 OpenAI 对函数名的字符限制
func SanitizeName(server, tool string) string {
	raw := server + "_" + tool
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}