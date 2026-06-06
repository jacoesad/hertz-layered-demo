# hz-server

CloudWeGo Hertz `hz` 示例，按接口入口拆成 console 和 openapi，按业务模块拆成 task 和 subtask。

## IDL namespace

```text
idl/console/task.thrift       namespace go console.task
idl/console/subtask.thrift    namespace go console.subtask
idl/openapi/task.thrift       namespace go openapi.task
idl/openapi/subtask.thrift    namespace go openapi.subtask
```

生成后的 `biz` 目录会跟 namespace 对齐：

```text
biz/handler/console/task
biz/handler/console/subtask
biz/handler/openapi/task
biz/handler/openapi/subtask

biz/model/console/task
biz/model/console/subtask
biz/model/openapi/task
biz/model/openapi/subtask
```

## 生成命令

```bash
hz new --module hz-server --idl idl/console/task.thrift
hz update --idl idl/console/subtask.thrift
hz update --idl idl/openapi/task.thrift
hz update --idl idl/openapi/subtask.thrift
```

## 分层

`biz` 按接口入口组织，`internal` 按业务领域组织：

```text
biz/handler/console/task    -> internal/task/service -> internal/task/repo -> internal/task/sqlstore -> SQLite
biz/handler/openapi/task    -> internal/task/service -> internal/task/repo -> internal/task/sqlstore -> SQLite

biz/handler/console/subtask.GetSubtask   -> internal/subtask/service -> main subtask repo      -> SQLite
biz/handler/console/subtask.ListSubtasks -> internal/subtask/service -> StarRocks subtask repo -> SQLite-simulated StarRocks
biz/handler/openapi/subtask              -> internal/subtask/service -> main subtask repo      -> SQLite
```

`console` 和 `openapi` 可以复用同一套内部业务能力，但各自拥有独立 IDL、路由、请求响应模型和 handler。

## 启动链路

```text
main.go
  -> config.Init("config/app.yaml")
      -> config.Load(...)
      -> validate config
      -> store config for optional config.Get()/MustGet()
  -> database.Init(ctx, cfg)
      -> database.Open(ctx, cfg.Database)
          -> sql.Open("sqlite3", "data/hz-server.db")
          -> migrate tables
          -> seed demo rows
      -> database.OpenStarRocks(ctx, cfg.StarRocks)
          -> sql.Open("sqlite3", "data/starrocks.db")
          -> migrate tables
          -> seed StarRocks demo rows
  -> app.NewContainer(dataSources)
      -> tasksqlstore.New(dataSources.DB)
      -> taskrepo.New(taskSQL)
      -> taskservice.New(taskRepo)
      -> subtasksqlstore.New(dataSources.DB)
      -> subtaskrepo.New(subtaskSQL)
      -> subtasksqlstore.New(dataSources.StarRocksDB)
      -> subtaskrepo.New(starRocksSubtaskSQL)
      -> subtaskservice.NewWithStarRocks(subtaskRepo, starRocksSubtaskRepo)
  -> app.Init(container)
  -> server.Default(server.WithHostPorts(cfg.Server.Addr))
  -> register(h)
  -> h.Spin()
```

`app.Init(container)` 会把 main 创建好的容器放到 `internal/app.Default`。hz 生成的 router 仍然调用包级 handler 函数，handler 在请求进来时通过 `app.MustDefault()` 取到 service。

## 请求链路

以控制台查询子任务为例：

```text
GET /console/v1/subtasks/:subtask_id?tenant_id=tenant-a
  -> biz/router/console/subtask
  -> biz/handler/console/subtask.GetSubtask
      -> BindAndValidate IDL request model
      -> app.MustDefault().SubtaskService
      -> internal/subtask/service.GetSubtask
      -> internal/subtask/repo.FindByTenantAndID
      -> internal/subtask/sqlstore.SelectByTenantAndID
      -> SQLite subtasks table
      -> repo.Row
      -> domain.Subtask
      -> biz/model/console/subtask.ConsoleSubtaskInfo
      -> JSON response
```

openapi 的 task/subtask 入口也是同样链路，只是在 handler 里映射成 openapi 自己的 response model，所以它可以少暴露一些字段。

DTO 转换放在各 handler 包的 `assembler.go` 中，例如：

```text
biz/handler/console/subtask/assembler.go
  domain.Subtask -> biz/model/console/subtask.ConsoleSubtaskInfo

biz/handler/openapi/subtask/assembler.go
  domain.Subtask -> biz/model/openapi/subtask.OpenAPISubtaskInfo
```

assembler 只做模型转换，不查数据库、不读配置、不放业务规则。

控制台列表接口 `GET /console/v1/subtasks?tenant_id=...` 特意走 StarRocks repo：

```text
ListSubtasks
  -> internal/subtask/service.ListSubtasks
  -> starRocksRepo.ListByTenant
  -> internal/subtask/sqlstore.SelectByTenant
  -> data/starrocks.db
```

## 配置和数据库

配置文件：

```text
config/app.yaml
```

当前配置：

```yaml
server:
  addr: ":8889"

database:
  driver: sqlite3
  dsn: data/hz-server.db

starrocks:
  driver: sqlite3
  dsn: data/starrocks.db
```

启动时 `main.go` 会读取配置，再初始化容器、打开 SQLite、执行建表，并写入示例数据：

```text
internal/config      读取配置
internal/database    打开主库和 StarRocks 模拟库、建表、seed，并管理连接关闭
internal/app         接收已初始化的数据源，装配 service/repo/sqlstore，并保存 Default 容器
```

本地数据库文件在 `data/hz-server.db` 和 `data/starrocks.db`，已加入 `.gitignore`。

可以用 `-config` 指定不同配置文件：

```bash
go run . -config config/app.yaml
```

## 示例接口

```bash
curl -s 'http://127.0.0.1:8889/console/v1/tasks/1001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/openapi/v1/tasks/1001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/console/v1/subtasks?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/console/v1/subtasks/5001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/openapi/v1/subtasks/5001?tenant_id=tenant-a'
```

console 响应包含 `tenant_id`、`owner`、`assignee` 等内部字段；openapi 响应暴露更少字段。
