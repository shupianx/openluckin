# 返回字段说明

来源：瑞幸官方工具文档 https://open.lkcoffee.com/docs（2026-06-10 整理），并经真实调用核对。

所有命令 stdout 输出统一 JSON 包络：

```json
{"code": 0, "msg": "success", "data": ...}
```

`code` 为 0 表示业务成功，非 0 时 `msg` 为错误信息。下文各节描述 `data` 部分。

## 公共结构

### 门店对象（query-shop-list 的 data 数组元素；preview-order / query-order-detail-info 的 shopInfo）

| 字段 | 类型 | 说明 |
|---|---|---|
| deptId | number | 门店ID（下游命令的 --dept-id 用它） |
| deptName | string | 门店名称 |
| address | string | 门店地址 |
| deptTags | array[string] | 门店标签 |
| longitude / latitude | number | 门店经纬度 |
| workTimeStart / workTimeEnd | string | 营业开始/结束时间 |
| distance | number | 距离，单位千米 |
| number | string | 门店编号，如 "(No.24404)" |
| workStatus | string | 营业状态，如 "营业中"（实测存在，官方文档未列） |

### 商品对象（search-product 的 data 数组元素；switch-product / query-product-detail-info 的 data）

| 字段 | 类型 | 说明 |
|---|---|---|
| productId | number | 商品ID |
| productName | string | 商品名称 |
| skuCode | string | 商品 SKU 编码（switch-product 返回的是切换后的新 skuCode） |
| pictureUrl | string | 商品图片URL |
| productAttrs | array[object] | 商品属性列表，见下 |
| tags | array[string] | 商品标签 |
| initialPrice | number | 面价 |
| estimatePrice | number | 预估到手价 |

productAttrs 每项：

| 字段 | 类型 | 说明 |
|---|---|---|
| attributeId | number | 属性组ID（如 杯型/温度/糖度） |
| attributeName | string | 属性组名称 |
| productSubAttrs | array[object] | 属性值列表：attributeId（属性值ID）、attributeName（属性值名称）、selected（boolean\|null 是否选中）、price（属性加价）、canSelected（number\|null 是否可选） |

## query-shop-list（queryShopList）

data：门店对象数组，按距离排序。

## search-product（searchProductForMcp）

data：商品对象数组（推荐结果，经实测确认为数组）。

## switch-product（switchProduct）

data：切换属性后的单个商品对象，用其返回的新 skuCode 下单。

入参补充（官方文档有、schema 里没写）：`attrOperationParam.subAttr.operation` 为操作类型，**选中传 3**。

## query-product-detail-info（queryProductDetailInfo）

data：单个商品对象（含全部可选属性，用于决定 switch-product 怎么切）。

## preview-order（previewOrder）

| 字段 | 类型 | 说明 |
|---|---|---|
| aboutTime | number | 预计取餐/送达时间戳 |
| discountPrice | number | 实际付款价/商品总价 |
| shopInfo | object | 门店对象 |
| productInfoList | array[object] | 商品信息列表，见下 |
| estimateTotalPrice | number | 预估总价 |
| couponCodeList | array[string] | 优惠券编码列表（传给 create-order 的 --coupon-code-list） |
| orderGranularCommodityList | array[object] | 商品粒度价格：commodityId、commodityCode、commodityName、payableMoney（应付）、payMoney（实付）、expressExpectTime（number\|null 配送预计送达时间） |
| privilegeMoney | number | 优惠金额 |
| totalInitialPrice | number | 商品总面价 |

productInfoList 每项：productId、skuCode、name、amount、additionDesc（附属备注，如规格描述）、bigPicUrl / breviaryPicUrl（string|null 图片）、initPrice（面价）、estimatePrice（预估到手价）。

## create-order（createOrder）

| 字段 | 类型 | 说明 |
|---|---|---|
| orderId | number | 订单ID |
| orderIdStr | string | 字符串订单ID（**query-order-detail-info 的 --order-id 用它**） |
| payOrderUrl | string | 微信支付 URL |
| payOrderQrCodeUrl | string | 支付二维码链接 |
| discountPrice | number | 实付款价/商品总价 |
| needPay | boolean | 是否需要支付（true 时需引导用户付款） |
| tradeNo | string\|null | 交易号 |
| description | string\|null | 描述信息 |
| businessNotifyUrl | string\|null | 业务通知地址 |
| subMchid | string\|null | 微信支付子商户号 |

## query-order-detail-info（queryOrderDetailInfo）

| 字段 | 类型 | 说明 |
|---|---|---|
| orderId | string | 订单ID |
| orderStatus | number | 订单状态：10=待付款，20=下单成功，30=制作中，60=等待取餐，80=已完成，100=已取消 |
| orderStatusName | string | 订单状态名称 |
| aboutTime | number | 预计取餐/送达时间戳 |
| takeMealTime | string | 实际取餐时间 |
| takeMealCodeInfo | object | **取餐码信息：code（取餐码）、takeOrderId（取餐单ID）** |
| shopInfo | object | 门店对象 |
| productInfoList | array[object]\|null | 商品信息列表（同 preview-order，待付款订单可能为 null） |
| orderPayAmount | number | 订单支付金额 |
| dispatchInfo | object | 配送信息：dispatcherName、dispatcherMobile、dispatchAboutTime、destinationDistance |
| orderCommodityList | array[object] | 订单商品：commodityId、commodityCode、commodityName、payableMoney、payMoney |
| orderType | string | 订单类型 |
| customerParams | object\|null | 自定义参数 |

## cancel-order（cancelOrder）

data：boolean，是否取消成功。
