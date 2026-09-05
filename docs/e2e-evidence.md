# E2E 验收记录

脚本：`docs/e2e.sh`（BASE_URL 默认 `http://127.0.0.1:8080`）

覆盖路径：管理端登录 → Agent 匹配 Top5 → 确认生成 sid → 小程序登录浏览 → 下单支付(mock) → 订单 `notified` + 工会分成报表 → 管理端退款（ledger 取消）。

最近一次运行：`refunded ok` / `OK`。
