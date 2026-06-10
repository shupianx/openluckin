package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yu/openluckin/internal/mcpclient"
)

// 本文件是生成命令的运行时：tools/gen 从 schema/tools.json 生成
// gen_commands.go 里的 genToolCommands 表，这里负责把表变成 cobra 命令。
// 改映射逻辑改这里，改工具清单重跑 go generate。

// paramKind 描述 schema 参数到 flag 的映射方式。
type paramKind int

const (
	kindString      paramKind = iota // string → --flag value
	kindInteger                      // integer → int64 flag
	kindNumber                       // number → float64 flag
	kindBool                         // boolean → bool flag
	kindStringSlice                  // 字符串数组 → 逗号分隔/可重复 flag
	kindJSON                         // 嵌套对象/数组 → JSON 字符串 flag
)

type toolParam struct {
	flag     string // kebab-case 的 flag 名
	key      string // MCP 工具参数的原始字段名
	kind     paramKind
	required bool
	desc     string
}

type toolCommand struct {
	use    string // 子命令名
	tool   string // MCP 工具名
	short  string
	params []toolParam
}

// newToolCommand 把一条 toolCommand 表项变成可执行的 cobra 命令。
func newToolCommand(tc toolCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   tc.use,
		Short: tc.short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			toolArgs := map[string]any{}
			for _, p := range tc.params {
				// 可选参数未显式传入时不发送，避免把零值当成业务值。
				if !p.required && !cmd.Flags().Changed(p.flag) {
					continue
				}
				v, err := paramValue(cmd, p)
				if err != nil {
					return err
				}
				toolArgs[p.key] = v
			}
			return withClient(cmd.Context(), func(ctx context.Context, c *mcpclient.Client) error {
				result, err := c.CallTool(ctx, tc.tool, toolArgs)
				if err != nil {
					return err
				}
				fmt.Println(result)
				return nil
			})
		},
	}
	// 保持命令表里的参数顺序（官方业务顺序），不让 cobra 按字母重排帮助文本。
	cmd.Flags().SortFlags = false
	for _, p := range tc.params {
		desc := p.desc
		switch p.kind {
		case kindString:
			cmd.Flags().String(p.flag, "", desc)
		case kindInteger:
			cmd.Flags().Int64(p.flag, 0, desc)
		case kindNumber:
			cmd.Flags().Float64(p.flag, 0, desc)
		case kindBool:
			cmd.Flags().Bool(p.flag, false, desc)
		case kindStringSlice:
			cmd.Flags().StringSlice(p.flag, nil, desc+"（逗号分隔或重复传参）")
		case kindJSON:
			cmd.Flags().String(p.flag, "", desc+"（JSON 字符串）")
		}
		if p.required {
			_ = cmd.MarkFlagRequired(p.flag)
		}
	}
	return cmd
}

// paramValue 按参数类型从 flag 中取值，转成 MCP 工具期望的 JSON 形态。
func paramValue(cmd *cobra.Command, p toolParam) (any, error) {
	switch p.kind {
	case kindString:
		return cmd.Flags().GetString(p.flag)
	case kindInteger:
		return cmd.Flags().GetInt64(p.flag)
	case kindNumber:
		return cmd.Flags().GetFloat64(p.flag)
	case kindBool:
		return cmd.Flags().GetBool(p.flag)
	case kindStringSlice:
		return cmd.Flags().GetStringSlice(p.flag)
	case kindJSON:
		raw, err := cmd.Flags().GetString(p.flag)
		if err != nil {
			return nil, err
		}
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return nil, fmt.Errorf("--%s 不是合法 JSON: %w", p.flag, err)
		}
		return v, nil
	}
	return nil, fmt.Errorf("未知参数类型: %d", p.kind)
}

func init() {
	for _, tc := range genToolCommands {
		rootCmd.AddCommand(newToolCommand(tc))
	}
}
