# iamv2

`iamv2` 是 [marmotedu/iam](https://github.com/marmotedu/iam) 的 OneX 规范化迁移版本。
项目保留了 iam 的核心业务逻辑（用户、密钥、Ladon 授权策略、授权审计），同时按照
OneX 规范重新组织了目录结构、依赖注入、日志、错误处理、配置加载、存储访问和服务生命周期。

模块路径：`github.com/onexdemo/iamv2`

## 功能概览

- **iam-apiserver**：用户、密钥、策略的 REST 管理接口，JWT / HTTP Basic 认证，
  管理员引导账号，Redis 缓存失效通知。
- **iam-authz-server**：基于 Ladon 的授权决策服务，内存缓存密钥与策略，
  通过 Redis Pub/Sub 实时刷新，授权审计异步写入 Redis。
- **iam-pump**：消费授权审计数据，支持 MongoDB、Kafka、Elasticsearch、InfluxDB、
  Prometheus、CSV、Syslog、Dummy 等多种后端。
- **iam-watcher**：基于 OneX watch 调度器的周期任务，负责审计数据清理与长期未登录账号禁用。
- **iamctl**：命令行客户端，支持 user / secret / policy / jwt / version / info / validate 等命令。

## 技术栈与规范

- Go 1.27、Gin、GORM v2、Cobra、Viper、Wire
- OneX 规范：
  - `cmd/<binary>` + `internal/<service>` + `pkg/api` 目录结构
  - `github.com/onexstack/onexstack/pkg/app` 应用框架
  - `pkg/options` 配置项、`pkg/errorsx` 错误、`log/slog` 结构化日志
  - `pkg/server` HTTP 服务、`pkg/watch` 周期任务、`pkg/store` 泛型存储
  - `pkg/token` JWT、`pkg/rid` 资源 ID、`pkg/options` MySQL/Redis/Mongo 连接
  - `google/wire` 编译期依赖注入
- 配置优先级：命令行 flag > 配置文件 > 默认值。

## 目录结构

```text
.
├── cmd/                        # 5 个可执行入口
│   ├── iam-apiserver/
│   ├── iam-authz-server/
│   ├── iam-pump/
│   ├── iam-watcher/
│   └── iamctl/
├── internal/
│   ├── apiserver/              # 管理服务：model / store / biz / handler / validation / conversion
│   ├── authz/                  # 授权服务：store / cache / authorization / analytics / handler
│   ├── pump/                   # 审计数据消费：analytics / storage / pumps / watcher
│   ├── watcher/                # 周期任务：clean / task watcher
│   ├── iamctl/                 # CLI 命令树与 REST 客户端
│   └── pkg/                    # errno / contextx / known / middleware / notifier
├── pkg/api/                    # 公共 REST API 类型
├── configs/                    # 各服务配置示例
├── build/docker/               # 镜像构建文件
├── scripts/                    # 构建脚本
├── third_party/                # 第三方 protobuf 依赖（脚手架保留）
├── Makefile
└── onexstack.yaml              # 脚手架项目定义
```

分层依赖方向固定为：`handler → biz → store → model`，`store` 只通过
`onexstack/pkg/store` 泛型抽象访问数据库。

## 构建

```bash
make deps        # 可选：安装 protoc / wire 等工具
go mod tidy
go generate ./...   # 生成 Wire 代码
make build          # 构建全部 5 个二进制到 _output/platforms/<os>_<arch>/
```

`iamv2` 的 REST API 使用手写 Go 类型定义，不包含 protobuf API，
因此 `make protoc` 是保留的 no-op 目标。

## 运行

依赖：

- MySQL 8.x（`iam-apiserver`、`iam-watcher`）
- Redis 7.x（`iam-apiserver` 通知、`iam-authz-server`、`iam-pump`）
- 可选：MongoDB / Kafka / Elasticsearch / InfluxDB（按需启用 pump 后端）

### 1. 启动 apiserver

```bash
_output/platforms/$(go env GOOS)_$(go env GOARCH)/iam-apiserver \
  --config configs/iam-apiserver.yaml
```

首次启动时会自动建表，并在没有任何管理员时创建引导管理员：

```text
username: admin
password: Admin@2020   # 请在生产环境中通过 --admin-password 或配置修改
```

健康检查：`GET http://127.0.0.1:8080/healthz`

### 2. 启动 authz server

```bash
_output/platforms/$(go env GOOS)_$(go env GOARCH)/iam-authz-server \
  --config configs/iam-authz-server.yaml
```

`authz` 使用管理员账号从 apiserver 的 `/v1/cache/secrets` 与 `/v1/cache/policies`
拉取全量快照并在内存中缓存；apiserver 在密钥或策略变更后通过 Redis 发布通知，
authz 收到通知后立即刷新缓存。

### 3. 启动 pump / watcher

```bash
_output/platforms/$(go env GOOS)_$(go env GOARCH)/iam-pump    --config configs/iam-pump.yaml
_output/platforms/$(go env GOOS)_$(go env GOARCH)/iam-watcher --config configs/iam-watcher.yaml
```

`iam-pump` 默认配置使用 MongoDB；只想验证链路时可把 `pump.pumps` 改为：

```yaml
pump:
  purge-delay: 10
  pumps:
    dummy:
      type: dummy
```

### 4. 使用 iamctl

```bash
BIN=_output/platforms/$(go env GOOS)_$(go env GOARCH)/iamctl

$BIN --config configs/iamctl.yaml user list
$BIN --config configs/iamctl.yaml secret create demo --description "demo key"
$BIN --config configs/iamctl.yaml policy create allow-read \
  '{"subjects":["users:foo"],"actions":["get"],"effect":"allow","resources":["resources:articles:<.*>"]}'
```

## REST API

| 方法 | 路径 | 说明 | 认证 |
|---|---|---|---|
| POST | `/login`、`/v1/login` | 用户名密码登录，返回 JWT | 公开 |
| POST | `/logout`、`/v1/logout` | 清除 JWT Cookie | 公开 |
| POST | `/refresh`、`/v1/refresh` | 刷新 JWT | Bearer |
| POST | `/v1/users` | 创建用户 | 公开 |
| GET | `/v1/users` | 用户列表 | 认证 |
| GET/PUT/DELETE | `/v1/users/:username` | 用户详情/更新/删除 | 认证 |
| PUT | `/v1/users/:username/change-password` | 修改自己的密码 | 认证 |
| POST/GET/DELETE | `/v1/secrets` | 创建/列表/批量删除密钥 | 认证 |
| GET/PUT/DELETE | `/v1/secrets/:name` | 密钥详情/更新/删除 | 认证 |
| POST/GET/DELETE | `/v1/policies` | 创建/列表/批量删除策略 | 认证 |
| GET/PUT/DELETE | `/v1/policies/:name` | 策略详情/更新/删除 | 认证 |
| GET | `/v1/cache/secrets` | 全量密钥快照 | 管理员 |
| GET | `/v1/cache/policies` | 全量策略快照 | 管理员 |
| POST | `/v1/authz` | Ladon 授权决策 | 密钥 JWT / Basic |

`/v1/authz` 请求体即 Ladon `Request`：

```json
{
  "subject": "users:foo",
  "action": "get",
  "resource": "resources:articles:1"
}
```

响应：

```json
{"allowed": true, "denied": false}
```

Ladon 的资源/主体/动作支持 `<regex>` 语法，例如 `resources:articles:<.*>`。

## 配置

- `configs/iam-apiserver.yaml`：监听、OTel、MySQL、Redis、JWT、引导管理员
- `configs/iam-authz-server.yaml`：监听、OTel、Redis、apiserver 快照地址、分析参数
- `configs/iam-pump.yaml`：watch、Redis、pump 后端列表
- `configs/iam-watcher.yaml`：watch、MySQL、清理/账号维护参数
- `configs/iamctl.yaml`：apiserver 地址、认证方式、TLS 选项

所有配置项都可以通过命令行 flag 覆盖，例如：

```bash
iam-apiserver --coredb.addr 127.0.0.1:3306 --coredb.database iam --redis.addr 127.0.0.1:6379
```

## 与原始 iam 的差异

迁移到 OneX 规范时做了以下有意调整：

1. **日志**：移除自定义 `pkg/log`，统一使用 `log/slog` 与 OneX 日志适配器。
2. **错误**：移除 `marmotedu/errors` 与 `internal/pkg/code`，统一使用
   `onexstack/pkg/errorsx`，错误响应为扁平 `{reason, message}` 并返回真实 HTTP 状态码。
3. **存储**：MySQL 访问改为 `onexstack/pkg/store` 泛型 store + `where` 条件；
   原 etcd store 未迁移（OneX 项目默认 MySQL）。
4. **服务框架**：`app`、`server`、`options`、`watch`、`token`、`rid` 全部使用 OneX 包；
   原 `internal/pkg/server`、`pkg/app`、`pkg/shutdown`、`pkg/storage` 已删除。
5. **认证**：JWT 使用 `onexstack/pkg/token`，不再依赖 `appleboy/gin-jwt`；
   仍支持 Bearer、HTTP Basic 与 `jwt` Cookie。
6. **authz 数据源**：原 iam 通过 gRPC `Cache` 服务拉取密钥/策略，
   现改为 apiserver 的管理员 REST 快照接口，避免维护额外的 gRPC 服务。
7. **动态 job**：原 `iam-pump` / `iam-watcher` 的自定义 cron/锁实现改为
   `onexstack/pkg/watch` 调度器 + 分布式锁。
8. **CLI**：`iamctl` 使用轻量 REST 客户端重写，未迁移原 `iamctl new`（代码生成器）
   与 `iamctl set db`（直接改库）两个辅助命令。

## 开发约定

- 所有对外错误使用 `errorsx`，错误源点只封装一次，上层用 `%w` 补充上下文。
- 日志使用扁平 key-value，错误键统一为 `error`，key 使用 snake_case。
- 不记录密码、密钥、Token 等敏感信息。
- HTTP 处理器通过 `core.HandleJSONRequest` / `core.HandleQueryRequest` 统一绑定与响应。
- 数据库访问只出现在 `store` 层，业务层不直接依赖 `gorm.DB`。
- 新增服务依赖通过 `wire.go` 声明并由 `go generate ./...` 生成。

## License

MIT
