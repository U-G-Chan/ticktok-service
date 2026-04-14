# TickTok 项目 K3s 集群部署指南

本文档旨在介绍将本项目的微服务与中间件（原基于 `docker-compose` 与 `go run`）完整迁移至 K3s (K3d) 集群的部署方案。此方案同样适用于未来生产环境的 Kubernetes 部署。

## 1. 架构概述

整个 Kubernetes 部署划分为两个独立的 Namespace：
1. **`infra`** 命名空间：用于部署各类中间件组件，包含：
   - MySQL (数据持久化基于 local-path PVC)
   - Redis
   - MinIO (数据持久化基于 local-path PVC)
   - Kafka & Etcd
   - Prometheus & Grafana (包含通过 ConfigMap 注入的配置)
2. **`ticktok`** 命名空间：用于部署核心业务微服务，包含：
   - 微服务组件：`gateway`, `user`, `content`, `message`, `chatbot`, `worker`。
   - 配置管理：通过 ConfigMap (`ticktok-config`) 和 Secret (`ticktok-secret`) 提取环境变量并批量注入 Pod。
   - 流量入口：通过 Kubernetes Ingress 将宿主机的 `localhost:8080/api/v1/*` 流量统一路由至 `gateway` 服务。

---

## 2. 前置要求

在本地开始部署之前，你需要确保已经安装以下工具：
- [Docker Desktop](https://www.docker.com/) 且处于运行状态
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [k3d](https://k3d.io/)
- [Task](https://taskfile.dev/) （可选，用于便捷执行构建命令）

---

## 3. 集群初始化

我们通过 `k3d` 创建一个包含 LoadBalancer（映射本地 `8080` 端口）的轻量级 K8s 集群。

执行以下命令创建集群：
```bash
k3d cluster create ticktok-cluster --api-port 127.0.0.1:6550 --servers 1 --agents 2 --port "8080:80@loadbalancer"
```
> **注意**：此处映射了宿主机 `8080` 端口至集群负载均衡器的 `80` 端口，这使得我们可以直接通过 `http://localhost:8080` 访问 Ingress。

检查节点是否就绪：
```bash
kubectl get nodes
```

---

## 4. 微服务镜像构建与导入

本项目采用统一的多阶段构建 `Dockerfile`（位于项目根目录）。你可以通过 `SERVICE_NAME` 参数动态指定需要构建的微服务入口。

为了方便操作，我们已在 `Taskfile.yaml` 中配置了快捷指令：

### 4.1 编译镜像
在项目根目录运行以下命令以构建全部 6 个微服务的 Docker 镜像：
```bash
task build-images
```
*(如果没有安装 Task 工具，你可以手动执行类似 `docker build --build-arg SERVICE_NAME=gateway -t ticktok-gateway:latest .` 的命令)*

### 4.2 导入镜像至 k3d 集群
因为 k3d 运行在 Docker 容器内部，它无法直接读取宿主机上的镜像缓存。必须通过以下命令将打好的镜像导入至集群内部：
```bash
task import-images
```

---

## 5. Kubernetes 资源部署

所有 Kubernetes 的资源配置文件均存放于 `deploy/k8s/` 目录下。

### 5.1 部署流程
同样可以使用 `Taskfile.yaml` 中的快捷命令一键部署：
```bash
task k8s-deploy
```

**或者手动按顺序应用：**
1. 创建命名空间：
   ```bash
   kubectl apply -f deploy/k8s/namespaces.yaml
   ```

2. 准备配置文件（如果是首次拉取项目）：
   由于涉及到敏感信息，`secret.yaml` 和 `config.yaml` 被 Git 忽略了。请先从模板文件复制出正式配置：
   ```bash
   cp deploy/k8s/infra/secret.yaml.template deploy/k8s/infra/secret.yaml
   cp deploy/k8s/ticktok/config.yaml.template deploy/k8s/ticktok/config.yaml
   ```
   复制完成后，请打开这两个新文件，将其中的 `<YOUR_...>` 占位符修改为你真实的密码/密钥信息。

3. 部署 `infra` 中间件及其持久化声明：
   ```bash
   kubectl apply -f deploy/k8s/infra/
   ```
4. 部署 `ticktok` 业务微服务、ConfigMap 及 Ingress：
   ```bash
   kubectl apply -f deploy/k8s/ticktok/
   ```

### 5.2 检查部署状态
可以通过以下命令观察 Pod 是否正常启动（`ImagePullBackOff` 错误通常说明没有执行 `import-images` 导入镜像）：
```bash
kubectl get pods -n infra
kubectl get pods -n ticktok
```

---

## 6. 服务验证与排障

### 6.1 接口连通性验证
一旦 `ticktok` 命名空间下的 `gateway` Pod 处于 `1/1 Running` 状态，即可使用 curl 测试外部路由与服务的连通性：

```bash
# 测试开放的 feed 流接口
curl.exe -s http://localhost:8080/api/v1/feed

# 测试用户注册（涉及 Gateway -> User gRPC -> MySQL）
curl.exe -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d '{"username": "testuser_k8s", "password": "123456"}'
```

### 6.2 微服务探针 (Probes) 设计
- **gRPC 服务**（User, Content 等）：采用了 `tcpSocket` 探针检查端口监听状态。
- **HTTP 网关**（Gateway）：采用了 `tcpSocket` 检查 `8080` 端口监听状态（因为原本的 `/ping` 接口受 JWT 保护，直接 HTTP 检查会收到 `401 Unauthorized` 导致探针失效）。
- 如果 Pod 频繁重启（`CrashLoopBackOff`），通常是服务没有成功连接到内部的 `mysql` 或 `redis`，请查看日志。

### 6.3 查看日志与调试
```bash
# 查看网关日志
kubectl logs -n ticktok -l app=gateway

# 查看特定微服务日志
kubectl logs -n ticktok -l app=user

# 临时端口转发排查数据库（例如将集群 MySQL 转发到本地 3306）
kubectl port-forward -n infra svc/mysql 3306:3306
```

---

## 7. 配置覆盖与扩展说明
- 若在生产环境下，你只需要更新 `deploy/k8s/ticktok/config.yaml` 中的环境变量（特别是 `ticktok-secret` 里的密码与 Key）。
- 当前持久化方案依赖 K3s 自带的 `local-path` provisioner，生产上可平滑切换至 Ceph、EBS 或云厂商的 CSI StorageClass，只需修改 `PVC` 定义即可。
