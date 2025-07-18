# HavenCamp Docker部署指南

## 📋 简介

本文档描述如何使用Docker容器化部署HavenCamp聊天系统。通过Docker部署，您可以快速启动包含所有依赖服务的完整聊天系统。

## 🛠️ 前置要求

- Docker (版本 >= 20.10)
- Docker Compose (版本 >= 1.29)
- 8GB+ 可用内存
- 10GB+ 可用磁盘空间

## 🚀 快速启动

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd HavenCamp
```

### 2. 使用启动脚本

```bash
chmod +x docker-start.sh
./docker-start.sh
```

### 3. 手动启动（可选）

如果您不想使用启动脚本，可以手动执行：

```bash
# 创建必要目录
mkdir -p static/avatars static/files

# 启动所有服务
docker-compose up --build -d

# 查看服务状态
docker-compose ps
```

## 📦 服务架构

Docker部署包含以下服务：

| 服务 | 端口 | 说明 |
|------|------|------|
| frontend | 80 | Vue.js前端应用 |
| backend | 8000 | Go后端API服务 |
| mysql | 3306 | MySQL数据库 |
| redis | 6379 | Redis缓存 |
| kafka | 9092 | Kafka消息队列 |
| zookeeper | 2181 | Zookeeper服务 |

## 🔧 配置说明

### 数据库配置

默认MySQL配置：
- 数据库名: `haven_camp_server`
- 用户名: `root`
- 密码: `123456`
- 端口: `3306`

### Redis配置

默认Redis配置：
- 端口: `6379`
- 无密码

### 阿里云短信配置

如需使用短信功能，请修改 `configs/config.docker.toml` 文件：

```toml
[authCodeConfig]
accessKeyID = "your accessKeyID in alibaba cloud"
accessKeySecret = "your accessKeySecret in alibaba cloud"
signName = "阿里云短信测试"
templateCode = "SMS_154950909"
```

## 🌐 访问地址

启动成功后，您可以通过以下地址访问：

- **前端应用**: http://localhost
- **后端API**: http://localhost:8000
- **MySQL**: localhost:3306
- **Redis**: localhost:6379
- **Kafka**: localhost:9092

## 📋 常用命令

### 查看服务状态
```bash
docker-compose ps
```

### 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f mysql
```

### 停止服务
```bash
docker-compose down
```

### 重启服务
```bash
docker-compose restart
```

### 重新构建并启动
```bash
docker-compose up --build -d
```

## 🔍 故障排除

### 1. 端口冲突

如果遇到端口冲突，请修改 `docker-compose.yml` 中的端口映射：

```yaml
services:
  frontend:
    ports:
      - "8080:80"  # 将前端端口改为8080
```

### 2. 内存不足

如果遇到内存不足，可以：
- 增加Docker内存限制
- 暂时禁用Kafka服务（修改backend依赖）

### 3. 数据库连接失败

检查MySQL服务状态：
```bash
docker-compose logs mysql
```

如果MySQL启动失败，可能需要：
- 检查磁盘空间
- 清理Docker数据：`docker system prune`

### 4. 前端无法访问后端

确保nginx配置正确，检查：
- `web/chat-server/nginx.conf` 中的代理配置
- 网络连接是否正常

## 📊 性能优化

### 1. 资源限制

在生产环境中，建议为每个服务设置资源限制：

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### 2. 数据持久化

重要数据已通过Docker卷进行持久化：
- MySQL数据: `mysql_data` 卷
- Redis数据: `redis_data` 卷
- 静态文件: `./static` 目录映射

### 3. 日志管理

配置日志轮转以避免日志文件过大：

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## 🔒 安全配置

### 1. 修改默认密码

在生产环境中，请修改：
- MySQL root密码
- Redis密码（如需要）
- JWT密钥（如果使用）

### 2. 网络安全

- 使用防火墙限制端口访问
- 配置SSL/TLS证书
- 使用Docker网络隔离

## 📝 开发模式

如果您需要开发模式，可以：

1. 使用卷映射源代码：
```yaml
volumes:
  - ./:/app
```

2. 启用热重载：
```yaml
command: go run main.go
```

## 🆘 获取帮助

如果您遇到问题，请：

1. 查看日志：`docker-compose logs -f`
2. 检查服务状态：`docker-compose ps`
3. 查看Docker系统状态：`docker system df`
4. 提交Issue到项目仓库

## 📚 相关链接

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose文档](https://docs.docker.com/compose/)
- [HavenCamp项目说明](./README.md) 