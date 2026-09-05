# 本地生活兴趣搭子

Go 单体后端 + MySQL + 管理端 Web + 微信小程序原生，实现：

客服在管理端 Agent 输入自然语言诉求 → AI 匹配 Top5 搭子 → 确认生成小程序链接 → 用户卡片浏览下单微信支付 → 工会分成记账并通知搭子。

## 目录

```
apps/admin-web     # Vue3 + Element Plus 管理端
apps/miniapp       # 微信小程序原生
services/api       # Go API（Gin + GORM + MySQL）
docs/              # 接口与方案说明
docker-compose.yml
```

## 快速启动

### 依赖

- Go 1.22+
- MySQL 8 / MariaDB 10.11+
- Node 20+（管理端）

### 数据库

```bash
mysql -e "CREATE DATABASE partner DEFAULT CHARACTER SET utf8mb4;
CREATE USER 'partner'@'%' IDENTIFIED BY 'partner';
GRANT ALL ON partner.* TO 'partner'@'%'; FLUSH PRIVILEGES;"
```

### API

```bash
cd services/api
cp configs/config.example.yaml configs/config.yaml
go run ./cmd/api -config configs/config.yaml
```

默认管理账号：`admin` / `admin123`  
LLM / 微信支付 / 微信登录默认 mock，可在配置中关闭。

### 管理端

```bash
cd apps/admin-web
npm install
npm run dev
```

### 小程序

用微信开发者工具打开 `apps/miniapp`，将 `utils/request.js` 中的 `API_BASE` 指向 API 地址。

### Docker Compose

```bash
docker compose up --build
```

- API: http://localhost:8080  
- 管理端: http://localhost:5173  

## 核心流程

1. 管理端登录 → Agent 匹配台输入诉求  
2. `POST /admin/match/chat` 解析并返回 Top5  
3. `POST /admin/match/:id/confirm` 生成 `sid` 与小程序链接  
4. 用户打开小程序 `/pages/discover/index?sid=...` 卡片浏览  
5. 询价下单 → 模拟/真实微信支付 → 工会分成 ledger + 搭子通知  

详见 [docs/API.md](docs/API.md)。
