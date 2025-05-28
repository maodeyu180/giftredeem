# GiftRedeem - 私密福利分发平台

GiftRedeem 是一个基于 Go 语言的平台，用于通过安全的 OAuth 认证系统分发私密福利（如兑换码、礼品卡等）。该平台允许用户创建、分享和兑换使用私密链接的福利。

## 核心功能

### 多渠道 OAuth 认证
- 支持各种 OAuth 提供商（LinuxDo、GitHub、Google 等）
- 跨多个 OAuth 账户的统一用户身份
- 账户绑定和合并

### 私密福利管理
- 创建包含多个兑换码的福利
- 自动代码验证和去重
- 带有唯一 UUID 标识符的私密分享链接
- 基于账户类型、账龄和提供商的领取限制

### 用户管理
- 双重用户角色：发布者和领取者
- OAuth 提供商管理和配置
- 用户资料和已领取福利跟踪

## 技术栈

- **后端**：Go 语言与 Gin 框架
- **数据库**：MySQL 与 GORM ORM
- **认证**：OAuth 2.0 与 JWT 令牌
- **API**：RESTful JSON API

## 项目结构

```
├── cmd/
│   └── server/         # 主应用程序入口点
├── config/             # 配置文件
├── internal/           # 内部应用程序代码
│   ├── api/            # API 处理程序
│   ├── auth/           # 认证逻辑
│   ├── benefit/        # 福利管理
│   ├── db/             # 数据库连接
│   ├── middleware/     # API 中间件
│   ├── models/         # 数据模型
│   └── utils/          # 实用函数
└── go.mod              # Go 模块定义
```

## 设置

### 前提条件

- Go 1.16 或更高版本
- MySQL 5.7 或更高版本

### 配置

基于提供的 `.env.example` 创建 `.env` 文件：

```env
# 服务器配置
PORT=8080

# 数据库配置
DB_USERNAME=root
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=giftredeem

# JWT 密钥
JWT_SECRET=your-secure-random-string

# OAuth 配置
OAUTH_LINUXDO_CLIENT_ID=your-client-id
OAUTH_LINUXDO_CLIENT_SECRET=your-client-secret
OAUTH_LINUXDO_AUTH_URL=https://connect.linux.do/oauth2/authorize
OAUTH_LINUXDO_TOKEN_URL=https://connect.linux.do/oauth/token
OAUTH_LINUXDO_USER_INFO_URL=https://connect.linux.do/api/user
```

### 数据库设置

1. 创建 MySQL 数据库：
   ```sql
   CREATE DATABASE giftredeem CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. 程序运行后 插入l sso 平台
```
INSERT INTO o_auth_providers (
    name,
    display_name,
    client_id,
    client_secret,
    auth_url,
    token_url,
    user_info_url,
    scope,
    enabled,
    sort_order,
    created_at
) VALUES (
    'linuxdo',
    'LinuxDo',
    'id',
    'secret',
    'https://connect.linux.do/oauth2/authorize',
    'https://connect.linux.do/oauth2/token',
    'https://connect.linux.do/api/user',
    'user',
    1,
    10,
    NOW()
);
```

```bash
go run ./cmd/server
```

服务器将在配置的端口上启动（默认：8080）。

## API 端点

### 认证

- `GET /api/auth/providers` - 获取可用的 OAuth 提供商
- `GET /api/auth/login/:provider` - 启动 OAuth 登录
- `GET /api/auth/callback/:provider` - OAuth 回调 URL
- `GET /api/auth/profile` - 获取当前用户资料

### 福利

- `POST /api/benefits` - 创建新福利
- `GET /api/benefits/my` - 获取当前用户创建的福利
- `PUT /api/benefits/:uuid/status` - 更新福利状态
- `GET /api/benefits/:uuid/claims` - 获取特定福利的领取记录

### 领取

- `GET /api/claims/my` - 获取当前用户领取的福利
- `GET /api/claim/:uuid` - 通过 UUID 查看福利
- `POST /api/claim/:uuid` - 领取福利

## 部署

### 前端部署
1. 构建 Vue 应用：
   ```bash
   cd vueweb/redeem
   npm run build
   ```

2. 将构建文件复制到前端目录：
   ```bash
   mkdir -p frontend/dist
   cp -r vueweb/redeem/dist/* frontend/dist/
   ```

### 使用 Nginx

配置 Nginx 作为反向代理：

```nginx
server {
    listen 80;
    server_name you domain;

    # 反向代理到 Go 服务
    location / {
        proxy_pass http://localhost:16666;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持（如果需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # 提高安全性的头部
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options SAMEORIGIN;
    add_header X-XSS-Protection "1; mode=block";
}
```

## 许可证

MIT 许可证 