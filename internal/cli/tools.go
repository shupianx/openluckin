package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yu/openluckin/internal/mcpclient"
)

// tools 子命令：列出服务端真实暴露的 MCP 工具，
// 用于摸清官方工具清单，也是连通性的冒烟测试。
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "列出瑞幸 MCP 服务暴露的所有工具及其参数 schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withClient(cmd.Context(), func(ctx context.Context, c *mcpclient.Client) error {
			tools, err := c.ListTools(ctx)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(tools, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		})
	},
}

// call 子命令：通用逃生通道，直接按工具名 + JSON 参数调用，
// 在高层子命令尚未覆盖某个工具时也能用。
var callCmd = &cobra.Command{
	Use:   "call <tool-name>",
	Short: "直接调用任意 MCP 工具：openluckin call <tool> --args '{\"k\":\"v\"}'",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawArgs, err := cmd.Flags().GetString("args")
		if err != nil {
			return err
		}
		toolArgs := map[string]any{}
		if rawArgs != "" {
			if err := json.Unmarshal([]byte(rawArgs), &toolArgs); err != nil {
				return fmt.Errorf("--args 不是合法 JSON: %w", err)
			}
		}
		return withClient(cmd.Context(), func(ctx context.Context, c *mcpclient.Client) error {
			result, err := c.CallTool(ctx, args[0], toolArgs)
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		})
	},
}

func init() {
	callCmd.Flags().String("args", "", "工具参数，JSON 对象字符串")
	rootCmd.AddCommand(toolsCmd, callCmd)
}
