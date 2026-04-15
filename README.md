# gin-demo

一个最小可用的 [`Golang`](go.mod) + [`Gin`](go.mod) + [`Gorm`](go.mod) 快速开发模板，默认使用纯 Go 驱动的 [`SQLite`](go.mod)，内置：

- YAML 配置加载
- 控制台 + 文件双写日志
- 按天或按小时切分日志文件
- 数据库初始化与自动迁移
- 用户 CRUD 示例
- 基础健康检查接口

## 目录结构

```text
cmd/server/main.go              程序入口
configs/config.sample.yaml      示例配置
internal/bootstrap/app.go       应用装配
internal/config/config.go       配置定义与加载
internal/model/user.go          数据模型
internal/dto/user.go            请求 DTO
internal/repository/user_repository.go  数据访问层
internal/service/user_service.go        业务层
internal/handler/user_handler.go        HTTP 处理层
internal/middleware/logger.go   审计日志中间件
pkg/logger/logger.go            日志组件
pkg/response/response.go        统一响应
```

## 快速开始

1. 安装依赖

```bash
go mod tidy
```

2. 直接运行

```bash
go run ./cmd/server -config configs/config.sample.yaml
```

3. 健康检查

```bash
curl http://127.0.0.1:8080/healthz
```

## 用户 CRUD 接口

- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

### 创建用户示例

```bash
curl -X POST http://127.0.0.1:8080/api/v1/users ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\"}"
```

### 查询用户列表示例

```bash
curl http://127.0.0.1:8080/api/v1/users
```

## 配置说明

[`configs/config.sample.yaml`](configs/config.sample.yaml) 默认内容：

- `server.port`：服务端口
- `server.ssl.enabled`：是否启用 HTTPS
- `log.dir`：日志目录
- `log.rotate_by`：`day` 或 `hour`
- `log.retention_days`：日志文件保留天数，默认 `7`
- `database.driver`：当前支持 `sqlite`
- `database.dsn`：SQLite 文件路径

## 日志输出

- 控制台实时输出
- 文件默认输出到 [`logs`](logs)
- 默认按天切分，例如 `app-20260415.log`
- 若将 `log.rotate_by` 改为 `hour`，则按小时切分
- 自动删除超过 `log.retention_days` 的历史日志文件
- 审计日志会记录请求方法、路径、Query、状态码、耗时、客户端 IP、完整请求体与完整响应体
