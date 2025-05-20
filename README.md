# Ticktok 聊天服务

这是一个使用Go语言开发的聊天应用后端服务，提供了好友管理、消息收发等功能。

## 功能特性

- 获取好友列表
- 获取消息列表
- 获取聊天历史记录
- 发送消息
- 标记消息为已读
- 获取用户信息
- 批量获取用户信息

## 技术栈

- Go 1.20+
- Gin Web框架
- GORM ORM框架
- MySQL数据库

## 项目结构

```
ticktok-service
  ├── config/               # 配置文件
  ├── internal/             # 内部代码包
  │   ├── handler/          # HTTP处理器
  │   ├── middleware/       # 中间件
  │   ├── model/            # 数据模型
  │   ├── pkg/              # 内部通用包
  │   ├── repository/       # 数据访问层
  │   └── service/          # 业务逻辑层
  ├── pkg/                  # 外部可导入的包
  │   └── util/             # 工具函数
  ├── config.yaml           # 配置文件
  ├── go.mod                # Go模块定义
  ├── go.sum                # 依赖校验和
  ├── main.go               # 主函数
  └── README.md             # 项目说明
```

## 快速开始

### 前置要求

- Go 1.20或更高版本
- MySQL 5.7或更高版本

### 安装

1. 克隆仓库

```bash
git clone https://github.com/yourusername/ticktok-service.git
cd ticktok-service
```

2. 安装依赖

```bash
go mod download
```

3. 配置数据库

修改`config.yaml`中的数据库连接信息。

4. 编译和运行

```bash
go build
./ticktok-service
```

## API文档

### 获取好友列表

```
GET /api/friends
```

### 获取消息列表

```
GET /api/messages
```

### 获取聊天历史记录

```
GET /api/chat/:userId
```

### 发送消息

```
POST /api/chat/send
```

### 标记消息为已读

```
PUT /api/chat/read/:userId
```

### 获取用户信息

```
GET /api/user/:userId
```

### 批量获取用户信息

```
POST /api/users/batch
```

## 数据库设计

项目使用以下数据表：

- users: 用户表
- friendships: 好友关系表
- messages: 消息表
- sessions: 消息会话表
- unread_messages: 未读消息表

## 许可证

MIT 