package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yu/openluckin/internal/config"
	"github.com/yu/openluckin/internal/mcpclient"
)

// login 实现官方 CLI 同款的 localhost 回调登录：
// 本地起临时 HTTP 服务 → 呼出浏览器登录 → 网页把 token 回传到
// 127.0.0.1:<port>/callback → 校验 cli_session → 写入 ~/.openluckin/.env。
const loginBaseURL = "https://open.lkcoffee.com/cli"

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "呼出浏览器登录瑞幸开放平台，自动获取并保存 token",
	RunE: func(cmd *cobra.Command, _ []string) error {
		port, _ := cmd.Flags().GetInt("port")
		wait, _ := cmd.Flags().GetDuration("wait")
		debug, _ := cmd.Flags().GetBool("debug")
		return runLogin(cmd.Context(), port, wait, debug)
	},
}

func init() {
	loginCmd.Flags().Int("port", 0, "本地回调端口，0 表示随机空闲端口")
	loginCmd.Flags().Duration("wait", 5*time.Minute, "等待浏览器登录完成的最长时间")
	loginCmd.Flags().Bool("debug", false, "打印回调收到的原始数据（适配排查用）")
	rootCmd.AddCommand(loginCmd)
}

type loginResult struct {
	token string
	err   error
}

func runLogin(ctx context.Context, port int, wait time.Duration, debug bool) error {
	session := randomHex(16)

	ln, actualPort, err := listenLoopback(port)
	if err != nil {
		return fmt.Errorf("无法监听本地回调端口: %w", err)
	}
	defer ln.Close()

	q := url.Values{}
	q.Set("auth", "login")
	q.Set("cli_session", session)
	q.Set("redirect_url", fmt.Sprintf("http://127.0.0.1:%d/callback", actualPort))
	loginURL := loginBaseURL + "?" + q.Encode()

	ch := make(chan loginResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// 同时兼容「页面 302 跳转回调」和「页面 fetch 回调」两种实现，
		// 后者需要 CORS 头。
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if debug {
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
			fmt.Printf("← 收到回调: %s %s\n", r.Method, r.URL.String())
			if ct := r.Header.Get("Content-Type"); ct != "" {
				fmt.Println("  Content-Type: " + ct)
			}
			if len(body) > 0 {
				fmt.Println("  Body: " + string(body))
			}
		}

		token, err := extractToken(r, session)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err != nil {
			fmt.Fprintf(w, resultHTML, "登录失败", err.Error())
			select {
			case ch <- loginResult{err: err}:
			default:
			}
			return
		}
		fmt.Fprintf(w, resultHTML, "登录成功", "token 已保存，请关闭此页面回到终端。")
		select {
		case ch <- loginResult{token: token}:
		default:
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	fmt.Println("正在打开浏览器完成登录，如未自动打开请手动访问：")
	fmt.Println("  " + loginURL)
	openBrowser(loginURL)

	var res loginResult
	select {
	case res = <-ch:
	case <-time.After(wait):
		return fmt.Errorf("等待登录回调超时（%s），可用 --wait 延长", wait)
	case <-ctx.Done():
		return ctx.Err()
	}
	if res.err != nil {
		return res.err
	}

	path, err := config.SaveToken(res.token)
	if err != nil {
		return fmt.Errorf("token 获取成功但保存失败: %w", err)
	}
	fmt.Println("✓ 登录成功，token 已保存到 " + path)
	warnLocalEnvOverride()
	verifyToken(ctx, res.token)
	return nil
}

// extractToken 从回调请求中提取 token，并校验 cli_session 防伪造。
// 官方实测格式为 POST JSON：{"token":"...","cli_session":"..."}；
// 这里同时保留 GET 查询串与 POST 表单的解析以防官方实现变化。
func extractToken(r *http.Request, wantSession string) (string, error) {
	params := r.URL.Query()
	if r.Method == http.MethodPost {
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				for k, v := range body {
					if s, ok := v.(string); ok {
						params.Set(k, s)
					}
				}
			}
		} else if err := r.ParseForm(); err == nil {
			for k, vs := range r.PostForm {
				if len(vs) > 0 {
					params.Set(k, vs[0])
				}
			}
		}
	}

	if got := params.Get("cli_session"); got != "" && got != wantSession {
		return "", fmt.Errorf("cli_session 不匹配，疑似伪造回调，已拒绝")
	}
	token := params.Get("token")
	if token == "" {
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		return "", fmt.Errorf("回调中未找到 token，收到的参数名为 %v，需要适配", keys)
	}
	return token, nil
}

// verifyToken 用新 token 实际连一次 MCP，确认登录拿到的 token 可用。
func verifyToken(ctx context.Context, token string) {
	cfg := config.Load()
	cfg.Token = token
	client := mcpclient.New(cfg)
	if err := client.Connect(ctx); err != nil {
		fmt.Println("⚠ token 已保存，但连通性验证失败：" + err.Error())
		return
	}
	defer client.Close()
	tools, err := client.ListTools(ctx)
	if err != nil {
		fmt.Println("⚠ token 已保存，但连通性验证失败：" + err.Error())
		return
	}
	fmt.Printf("✓ token 验证通过（已连通 MCP，共 %d 个工具）\n", len(tools))
}

// warnLocalEnvOverride 提醒：当前目录 .env 优先级更高，可能盖掉刚保存的 token。
func warnLocalEnvOverride() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	s := string(data)
	if strings.Contains(s, config.EnvToken+"=") || strings.Contains(s, config.EnvTokenAlias+"=") {
		fmt.Println("⚠ 注意：当前目录的 .env 里也配置了 token，且优先级高于 ~/" +
			config.ConfigDirName + "/.env，在本目录运行时生效的仍是它")
	}
}

// listenLoopback 绑定本地回调端口，port 为 0 时由系统分配空闲端口。
func listenLoopback(port int) (net.Listener, int, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, 0, err
	}
	return ln, ln.Addr().(*net.TCPAddr).Port, nil
}

func openBrowser(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("（自动打开浏览器失败，请手动访问上面的链接）")
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

const resultHTML = `<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>openluckin</title></head>
<body style="font-family:system-ui;display:flex;justify-content:center;margin-top:20vh">
<div style="text-align:center"><h2>%s</h2><p>%s</p></div>
</body></html>`
