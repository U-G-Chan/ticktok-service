# Ticktok 微服务后端

这是一个基于 Go 语言开发的短视频社交平台后端服务，采用现代化的微服务架构设计。项目拆分为多个独立的服务，支持高并发、高可用，并集成了 AI 智能助手功能。

## 项目架构

### 微服务拆分

- **Gateway (网关服务)**: 对外暴露 RESTful API，集成 JWT 鉴权，负责请求路由和协议转换 (HTTP -> gRPC)。
- **User (用户服务)**: 处理用户注册、登录、个人信息管理以及关注/粉丝等社交关系。
- **Content (内容服务)**: 负责视频流、视频发布、点赞等核心内容业务。
- **Message (消息服务)**: 提供私信聊天功能 (后续集成 WebSocket)。
- **Chatbot (AI 服务)**: 集成 LLM (如 Ollama/OpenAI)，提供智能问答助手功能。

### 技术栈

- **编程语言**: Go 1.24+
- **Web 框架**: Gin (HTTP), gRPC (RPC)
- **数据库**: MySQL 8.0 (GORM)
- **缓存**: Redis
- **对象存储**: MinIO (视频/图片存储)
- **服务发现**: Etcd
- **监控告警**: Prometheus + Grafana
- **配置管理**: Viper (支持 YAML/Env)
- **开发工具**: Air (热重载), Go-Task (任务管理)

### 目录结构

```
ticktok-service/
├── api/                  # Protobuf 定义文件 (gRPC 契约)
├── cmd/                  # 各微服务入口 (main.go)
├── config/               # 配置文件 (config.yaml, .air/*)
├── deploy/               # 部署相关 (docker-compose.yml)
├── internal/             # 业务逻辑核心代码
│   ├── gateway/          # 网关服务代码
│   ├── user/             # 用户服务代码
│   ├── content/          # 内容服务代码
│   ├── message/          # 消息服务代码
│   └── chatbot/          # AI 服务代码
├── pkg/                  # 公共工具库 (Logger, JWT, Errno 等)
└── Taskfile.yaml         # 任务管理脚本
```

## 快速开始

### 前置要求

- Go 1.24+
- Docker & Docker Compose
- Go-Task (可选，推荐)

### 1. 启动基础环境

使用 Docker Compose 一键启动 MySQL, Redis, MinIO, Etcd, Prometheus, Grafana。

```bash
# 使用 go-task
task docker-up

# 或者使用 PowerShell 脚本
.\run.ps1 docker-up
```

### 2. 生成 gRPC 代码

如果修改了 `api/` 下的 `.proto` 文件，需要重新生成 Go 代码：

```bash
task proto
# 或 .\run.ps1 proto
```

### 3. 运行微服务

#### 开发模式 (支持热重载)

```bash
task dev-gateway   # 启动网关
task dev-user      # 启动用户服务
task dev-content   # 启动内容服务
# ... 其他服务同理
```

#### 普通运行

```bash
task run-gateway
task run-user
# ...
```

## 基本功能

### 1. 用户体系 (User Service)

- [ ] 用户注册/登录 (JWT)
- [ ] 用户信息查询
- [ ] 关注/取消关注
- [ ] 获取粉丝/关注列表

### 2. 内容互动 (Content Service)

- [ ] 视频流推送 (Feed)
- [ ] 视频投稿/发布
- [ ] 视频点赞/取消点赞
- [ ] 视频评论列表/发表评论

### 3. 即时通讯 (Message Service)

- [ ] 发送私信
- [ ] 获取历史消息记录

### 4. 智能助手 (Chatbot Service)

- [ ] AI 智能问答接口

