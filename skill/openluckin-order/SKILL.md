---
name: openluckin-order
description: 通过 openluckin CLI 为用户点瑞幸咖啡。当用户想找瑞幸门店、搜索商品、下单买咖啡、查询或取消订单时使用。
---

# openluckin —— 瑞幸咖啡点单 CLI

> 本文档由 tools/gen 从官方 MCP 工具快照自动生成，请勿手改；
> 重新生成：make snapshot && go generate ./...

openluckin 把瑞幸官方 AI 开放平台的 MCP 点单服务封装成一次性命令行调用：
传参 → stdout 输出结果（JSON 文本）→ 退出。退出码 0 为成功；失败时退出码非 0，
错误信息输出到 stderr。命令之间无会话状态，可独立调用。

## 前置条件

- 可执行文件位置（不在 PATH 中，按平台区分）：
  - macOS / Linux："$HOME/.openluckin/bin/openluckin"
  - Windows（PowerShell）："$env:USERPROFILE\.openluckin\bin\openluckin.exe"
- 本文档示例均按 macOS/Linux 路径书写，Windows 下请替换为上述 exe 路径
- 可执行文件不存在时，替用户安装后继续（安装到上述固定位置，不要用其他方式）：
  - macOS / Linux：curl -fsSL https://openluckin.com/install.sh | bash
  - Windows（PowerShell）：irm https://openluckin.com/install.ps1 | iex
- 鉴权失败时替用户执行 login 子命令：它会自动呼出浏览器并阻塞等待，
  告知用户去浏览器完成登录；命令成功返回（token 自动保存）后重试原命令

## 标准点单流程

1. 用用户位置经纬度调 query-shop-list 找门店，从返回中取门店 ID（deptId）
2. 用 search-product 按自然语言搜索商品，从返回中取 productId 和 skuCode
3. 如需调整规格（冰量/糖度/杯型等），先 query-product-detail-info 查可选属性，
   再用 switch-product 切换属性，得到新的 skuCode
4. 用 preview-order 预览订单，向用户确认商品、规格、价格；返回中如有可用
   优惠券（couponCodeList），记下备用
5. 用户确认后调 create-order 正式下单（couponCodeList 传 preview 返回的值）
6. 若返回的 needPay 为 true：payOrderQrCodeUrl 本身就指向一张付款二维码
   图片，直接用 Markdown 图片语法即时展示给用户扫码：
   ![付款二维码](<payOrderQrCodeUrl>)
   无需用任何二维码生成工具转换，也不要只贴 URL 文本
7. 用户表示付款完成后，用 create-order 返回的 orderIdStr 调
   query-order-detail-info，确认 orderStatus 已不是 10（待付款）后，
   把 takeMealCodeInfo 中的取餐单ID（takeOrderId）和取餐码（code）告知
   用户；返回中如含取餐码二维码图片链接，同样用 Markdown 图片直接展示
8. 需要时用 cancel-order 取消订单

**重要约束**

- create-order 会产生真实订单和扣费，下单前必须向用户复述门店、商品、规格、
  数量、价格并获得明确同意
- 不要凭空构造 deptId / productId / skuCode，必须来自上游命令的真实返回
- 经纬度使用 GCJ-02 坐标系（国内地图通用坐标）

## 命令参考

各命令返回 JSON 的字段含义见 [references/responses.md](references/responses.md)，
解读返回结果（如提取 deptId、skuCode、支付链接、取餐码）前先查阅。

### query-shop-list — 瑞幸咖啡查询门店列表

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --longitude | number | ✓ | 经度 |
| --latitude | number | ✓ | 纬度 |
| --dept-name | string |  | 门店名称 |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" query-shop-list --longitude 116.3975 --latitude 39.9087
```

### search-product — 瑞幸咖啡根据用户传入query， 匹配商品推荐结果

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --dept-id | integer | ✓ | 门店ID |
| --query | string | ✓ | 用户原始查询文本 |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" search-product --dept-id 1234 --query '生椰拿铁'
```

### switch-product — 瑞幸咖啡商品属性切换

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --dept-id | integer | ✓ | 门店ID |
| --product-id | integer | ✓ | 商品ID |
| --sku-code | string | ✓ | 商品SKU编码 |
| --attr-operation-param | JSON | ✓ | 商品属性切换参数 |
| --amount | integer | ✓ | 商品数量 |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" switch-product --dept-id 1234 --product-id 5826 --sku-code 'SKU123' --attr-operation-param '{"attributeId":1,"subAttr":{"attributeId":2,"operation":1}}' --amount 1
```

--attr-operation-param 的取值结构（JSON Schema）：

```json
{
  "description": "商品属性切换参数",
  "properties": {
    "attributeId": {
      "format": "int64",
      "type": "integer"
    },
    "subAttr": {
      "properties": {
        "attributeId": {
          "format": "int64",
          "type": "integer"
        },
        "operation": {
          "format": "int32",
          "type": "integer"
        }
      },
      "required": [
        "attributeId",
        "operation"
      ],
      "type": "object"
    }
  },
  "required": [
    "attributeId",
    "subAttr"
  ],
  "type": "object"
}
```

### query-product-detail-info — 瑞幸咖啡查询商品详情

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --dept-id | integer | ✓ | 门店ID |
| --product-id | integer | ✓ | 商品ID |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" query-product-detail-info --dept-id 1234 --product-id 5826
```

### preview-order — 瑞幸咖啡订单预览

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --dept-id | integer | ✓ | 门店ID |
| --product-list | JSON | ✓ | 订单商品列表 |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" preview-order --dept-id 1234 --product-list '[{"productId":5826,"skuCode":"SKU123","amount":1}]'
```

--product-list 的取值结构（JSON Schema）：

```json
{
  "description": "订单商品列表",
  "items": {
    "properties": {
      "amount": {
        "format": "int32",
        "type": "integer"
      },
      "productId": {
        "format": "int64",
        "type": "integer"
      },
      "skuCode": {
        "type": "string"
      }
    },
    "required": [
      "amount",
      "productId",
      "skuCode"
    ],
    "type": "object"
  },
  "type": "array"
}
```

### create-order — 瑞幸咖啡创建订单

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --dept-id | integer | ✓ | 门店id |
| --product-list | JSON | ✓ | 订单商品列表 |
| --longitude | number | ✓ | 经度 |
| --latitude | number | ✓ | 纬度 |
| --coupon-code-list | string 列表 |  | 优惠券列表，此参数来自 `previewOrder` 的返回字段 `couponCodeList` |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" create-order --dept-id 1234 --product-list '[{"productId":5826,"skuCode":"SKU123","amount":1}]' --longitude 116.3975 --latitude 39.9087
```

--product-list 的取值结构（JSON Schema）：

```json
{
  "description": "订单商品列表",
  "items": {
    "properties": {
      "amount": {
        "format": "int32",
        "type": "integer"
      },
      "productId": {
        "format": "int64",
        "type": "integer"
      },
      "skuCode": {
        "type": "string"
      }
    },
    "required": [
      "amount",
      "productId",
      "skuCode"
    ],
    "type": "object"
  },
  "type": "array"
}
```

### query-order-detail-info — 瑞幸咖啡查询订单详情

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --order-id | string | ✓ | 订单ID |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" query-order-detail-info --order-id '202606100001'
```

### cancel-order — 瑞幸咖啡取消订单

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| --order-id | string | ✓ | 订单ID |

示例：

```bash
"$HOME/.openluckin/bin/openluckin" cancel-order --order-id '202606100001'
```

## 通用逃生通道

高层命令未覆盖的工具可直接透传调用：

```bash
"$HOME/.openluckin/bin/openluckin" tools                      # 列出底层全部工具及 schema
"$HOME/.openluckin/bin/openluckin" call <tool> --args '{...}' # 按工具名直接调用
```
