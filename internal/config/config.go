// Package config 负责加载 openluckin 的运行配置。
// 优先级：命令行 flag > 进程环境变量 > ./.env（开发） > ~/.openluckin/.env（生产） > 默认值。
// flag 的覆盖在 cli 层完成，这里处理其余三级。
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 默认对接瑞幸官方 AI 开放平台的订单 MCP 端点。
const DefaultEndpoint = "https://gwmcp.lkcoffee.com/order/user/mcp"

// ConfigDirName 是生产环境的配置目录（位于用户主目录下）。
const ConfigDirName = ".openluckin"

const (
	EnvEndpoint = "LUCKIN_MCP_ENDPOINT"
	EnvToken    = "LUCKIN_MCP_TOKEN"
	// EnvTokenAlias 兼容官方申请页使用的变量名。
	EnvTokenAlias = "LUCKIN_MCP_ORDER_TOKEN"
)

// Config 是连接瑞幸 MCP 服务所需的全部配置。
type Config struct {
	// Endpoint 是 MCP streamable HTTP 端点。
	Endpoint string
	// Token 是 Bearer 鉴权 token（在 open.lkcoffee.com 申请）。
	Token string
	// Timeout 是单次工具调用的超时时间。
	Timeout time.Duration
}

// Load 按 进程环境变量 > ./.env > ~/.openluckin/.env 的优先级合并配置。
func Load() Config {
	cfg := Config{
		Endpoint: DefaultEndpoint,
		Timeout:  30 * time.Second,
	}

	// 低优先级先写入，高优先级后写入覆盖。
	vars := map[string]string{}
	if home, err := os.UserHomeDir(); err == nil {
		mergeEnvFile(vars, filepath.Join(home, ConfigDirName, ".env"))
	}
	mergeEnvFile(vars, ".env")
	for _, k := range []string{EnvEndpoint, EnvToken, EnvTokenAlias} {
		if v := os.Getenv(k); v != "" {
			vars[k] = v
		}
	}

	if v := vars[EnvEndpoint]; v != "" {
		cfg.Endpoint = v
	}
	if v := vars[EnvToken]; v != "" {
		cfg.Token = v
	} else if v := vars[EnvTokenAlias]; v != "" {
		cfg.Token = v
	}
	return cfg
}

// mergeEnvFile 解析 KEY=VALUE 格式的 .env 文件并合并进 dst。
// 文件不存在时静默跳过；支持 # 注释、export 前缀和成对引号。
func mergeEnvFile(dst map[string]string, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
			v = v[1 : len(v)-1]
		}
		if k != "" {
			dst[k] = v
		}
	}
}

// SaveToken 把 token 写入 ~/.openluckin/.env（保留文件中的其他配置项），
// 返回写入路径。目录与文件不存在时自动创建，权限 0700/0600。
// 旧的 token 行（含别名变量）会被替换，避免新旧 token 并存。
func SaveToken(token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ConfigDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, ".env")

	newLine := EnvToken + "=" + token
	var lines []string
	replaced := false
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		for _, l := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			trimmed := strings.TrimPrefix(strings.TrimSpace(l), "export ")
			if strings.HasPrefix(trimmed, EnvToken+"=") || strings.HasPrefix(trimmed, EnvTokenAlias+"=") {
				if !replaced {
					lines = append(lines, newLine)
					replaced = true
				}
				continue
			}
			lines = append(lines, l)
		}
	}
	if !replaced {
		lines = append(lines, newLine)
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// Validate 在真正发起 MCP 调用前检查必填项。
func (c Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("MCP endpoint 为空，请通过 --endpoint 或 " + EnvEndpoint + " 设置")
	}
	if c.Token == "" {
		return errors.New("缺少鉴权 token：请在 ~/" + ConfigDirName + "/.env 或当前目录 .env 写入 " +
			EnvToken + "=<token>，或通过 --token / 环境变量传入（token 在 open.lkcoffee.com 申请）")
	}
	return nil
}
