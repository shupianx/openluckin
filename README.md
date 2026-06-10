# OpenLuckin ☕

> 一句话，幸运到手 —— 非官方瑞幸咖啡 CLI + AI Agent Skill，基于瑞幸官方 MCP 接口封装，让任何智能体帮你完成找店、点单、付款、取餐码全流程。

官网：[openluckin.com](https://openluckin.com)

## 这是什么

OpenLuckin 把瑞幸官方 AI 开放平台（[open.lkcoffee.com](https://open.lkcoffee.com)）的 MCP 点单服务封装成**一次性命令行调用**：传参 → 返回 JSON → 退出，无会话无状态。再配上一份写给 AI Agent 的 Skill 说明书，你的智能体就能端到端替你点咖啡。

## 使用方法

### 方式一：给 AI Agent 装 Skill（推荐）

把下面两行直接发给你的智能体（Claude Code 等支持 Skill 的 Agent）：

```
请下载安装 OpenLuckin Skill：
https://openluckin.com/openluckin-order.zip
```

也可以手动安装：下载 zip，解压到 Agent 的 skills 目录（如 `~/.claude/skills/`），得到 `openluckin-order/` 文件夹即可。

然后对它说：

```
帮我点一杯生椰拿铁
```

Agent 会按 Skill 说明书自动完成：

1. 检测 CLI 未安装时自动安装（无需手动操作）
2. 鉴权失败时呼出浏览器，引导你登录瑞幸账号（token 只存你本机）
3. 按你的位置找门店 → 搜商品 → 预览订单并**向你确认价格**
4. 下单后把微信支付链接渲染成二维码给你扫
5. 你说付款完成后，查询订单并报出**取餐码**

> ⚠️ 下单产生真实消费。Skill 中已写明 Agent 必须在 create-order 前向你复述门店、商品、价格并获得明确同意。

### 方式二：直接用 CLI

**安装**

```bash
# macOS / Linux
curl -fsSL https://openluckin.com/install.sh | bash

# Windows (PowerShell)
irm https://openluckin.com/install.ps1 | iex
```

二进制安装在 `~/.openluckin/bin/openluckin`（Windows 为 `%USERPROFILE%\.openluckin\bin\openluckin.exe`），脚本会自动加入 PATH 并校验 sha256。

**登录**

```bash
openluckin login
```

呼出浏览器登录瑞幸账号，token 自动保存到本机 `~/.openluckin/.env`（权限 600），之后所有命令免配置。

**点一杯咖啡的完整流程**

```bash
# 1. 用经纬度找门店（GCJ-02 坐标），拿到 deptId
openluckin query-shop-list --longitude 116.3975 --latitude 39.9087

# 2. 搜商品，拿到 productId 和 skuCode
openluckin search-product --dept-id 612691 --query '生椰拿铁'

# 3. 预览订单，确认价格与可用优惠券
openluckin preview-order --dept-id 612691 \
  --product-list '[{"productId":1262,"skuCode":"SP2077-00347","amount":1}]'

# 4. 下单（产生真实消费！），返回支付链接 payOrderUrl
openluckin create-order --dept-id 612691 --longitude 116.3975 --latitude 39.9087 \
  --product-list '[{"productId":1262,"skuCode":"SP2077-00347","amount":1}]'

# 5. 付款后查订单，takeMealCodeInfo 里是取餐码
openluckin query-order-detail-info --order-id '<orderIdStr>'
```

所有命令输出统一 JSON 包络 `{"code":0,"msg":"success","data":...}`，退出码 0 为成功。

## 命令一览

| 命令 | 作用 |
|---|---|
| `login` | 浏览器登录，自动获取并保存 token |
| `query-shop-list` | 按经纬度查附近门店 |
| `search-product` | 自然语言搜商品 |
| `query-product-detail-info` | 查商品详情与可选规格 |
| `switch-product` | 切换规格（冰量/糖度/杯型） |
| `preview-order` | 订单预览（价格 / 优惠券） |
| `create-order` | 创建订单（真实扣费） |
| `query-order-detail-info` | 查订单状态与取餐码 |
| `cancel-order` | 取消订单 |
| `tools` | 列出底层 MCP 工具及 schema |
| `call <tool>` | 按工具名直接透传调用 |

每个命令的参数用 `openluckin <命令> --help` 查看；返回字段说明见 [skill/openluckin-order/references/responses.md](skill/openluckin-order/references/responses.md)。

## 配置

加载优先级：命令行 flag > 环境变量 > `./.env` > `~/.openluckin/.env` > 默认值。

| 变量 | 说明 |
|---|---|
| `LUCKIN_MCP_TOKEN` | 鉴权 token（`login` 自动写入；兼容 `LUCKIN_MCP_ORDER_TOKEN`） |
| `LUCKIN_MCP_ENDPOINT` | MCP 端点，默认官方 `gwmcp.lkcoffee.com/order/user/mcp` |

token 只保存在你本机，不经过任何第三方服务器。

## 工作原理

```
AI Agent ──读 SKILL.md──> openluckin CLI ──MCP (streamable HTTP)──> 瑞幸官方 AI 开放平台
```

- CLI 的 8 个点单命令由官方 MCP 工具 schema **静态生成**（`schema/tools.json` 快照 → `go generate` → 命令表 + SKILL.md），官方接口变更时重新生成，代码与文档永远同步；
- 每次调用建立一条 MCP 会话、调一个工具、退出——无状态设计正是 Agent 能稳定调用的原因。

## 开发

需要 Go 1.26+、Node 22+。

```bash
make build      # 编译 CLI 到 bin/
make check      # build + vet + test
make snapshot   # 从官方 MCP 拉取工具清单快照（需 token）
make generate   # 由快照重新生成子命令表 + SKILL.md
make release    # 五平台交叉编译 + skill 打包，输出到 web/public/
```

## 免责声明

OpenLuckin 是个人开源项目，与瑞幸咖啡官方无关。底层调用的是瑞幸官方对外提供的 AI 开放平台接口，登录使用你自己的瑞幸账号。下单产生真实消费，请谨慎操作，使用产生的一切后果由使用者自行承担。
