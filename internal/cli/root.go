// Package cli 定义 openluckin 的命令行入口。
// 设计目标：外部 agent 只需按 SKILL.md 里的说明传参即可完成点单全流程。
package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/yu/openluckin/internal/config"
	"github.com/yu/openluckin/internal/mcpclient"
)

// version 由 release 构建时通过 -ldflags -X 注入。
var version = "dev"

var (
	flagEndpoint string
	flagToken    string
	flagTimeout  time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "openluckin",
	Short: "瑞幸咖啡非官方点单 CLI（基于官方 MCP 服务封装）",
	Long: `openluckin 把瑞幸官方 AI 开放平台的 MCP 点单服务封装成普通命令行，
外部 agent 通过子命令 + 参数即可完成搜门店、看菜单、下单等操作。

鉴权 token 在 https://open.lkcoffee.com 申请，
通过 --token 或环境变量 ` + config.EnvToken + ` 传入。`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
}

// Execute 是 main 的唯一入口。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagEndpoint, "endpoint", "", "MCP 端点（默认官方端点，亦可用 "+config.EnvEndpoint+"）")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Bearer 鉴权 token（亦可用 "+config.EnvToken+"）")
	rootCmd.PersistentFlags().DurationVar(&flagTimeout, "timeout", 0, "单次调用超时，如 30s")
}

// loadConfig 合并环境变量与 flag，flag 优先。
func loadConfig() config.Config {
	cfg := config.Load()
	if flagEndpoint != "" {
		cfg.Endpoint = flagEndpoint
	}
	if flagToken != "" {
		cfg.Token = flagToken
	}
	if flagTimeout > 0 {
		cfg.Timeout = flagTimeout
	}
	return cfg
}

// withClient 处理连接的建立与关闭，子命令只写业务逻辑。
func withClient(ctx context.Context, fn func(ctx context.Context, c *mcpclient.Client) error) error {
	cfg := loadConfig()
	client := mcpclient.New(cfg)
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()
	return fn(ctx, client)
}
