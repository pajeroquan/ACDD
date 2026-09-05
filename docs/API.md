# API 约定

统一响应：`{ "code": 0, "message": "ok", "data": ... }`  
管理端鉴权：`Authorization: Bearer <admin_jwt>`  
小程序鉴权：`Authorization: Bearer <user_jwt>`

## 管理端

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/login` | 登录 `{username,password}` |
| GET/POST | `/admin/unions` | 工会列表/创建 |
| PATCH | `/admin/unions/:id` | 更新工会（含 `commission_rate` 万分比） |
| GET/POST | `/admin/partners` | 搭子列表/创建 |
| GET/PUT | `/admin/partners/:id` | 搭子详情/更新 |
| POST | `/admin/match/chat` | Agent 匹配 `{text, match_request_id?}` |
| POST | `/admin/match/:id/confirm` | 确认候选 `{partner_ids}` → `{sid, mini_program_url}` |
| GET | `/admin/orders` | 订单列表 |
| GET | `/admin/orders/:id` | 订单详情 |
| GET | `/admin/commission/report` | 工会分成汇总 |
| GET | `/admin/notifications` | 搭子通知 inbox |

## 小程序

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/wx/login` | `{code, nickname}` → token |
| GET | `/api/browse/:sid` | 推荐卡片列表 |
| POST | `/api/orders/quote` | 询价 |
| POST | `/api/orders` | 下单 |
| POST | `/api/orders/:id/pay` | 发起支付（mock 时返回 `mock:true`） |
| POST | `/api/pay/mock-notify` | 模拟支付成功 `{out_trade_no}` |
| POST | `/api/pay/notify` | 微信支付回调（MVP JSON） |

## 分成规则

支付成功时读取搭子所属工会当前 `commission_rate`（万分比），快照写入订单与 `commission_ledgers`。无工会则分成=0。
