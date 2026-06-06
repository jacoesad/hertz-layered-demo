# hz-server

CloudWeGo Hertz `hz` 示例，按接口入口拆成 console 和 openapi，按业务模块拆成 task 和 subtask。

## IDL namespace

```text
idl/console/task.thrift       namespace go console.task
idl/console/subtask.thrift    namespace go console.subtask
idl/openapi/task.thrift       namespace go openapi.task
idl/openapi/subtask.thrift    namespace go openapi.subtask
idl/downstream/task_runner.thrift namespace go downstream.taskrunner
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

`idl/downstream/task_runner.thrift` 用于描述当前服务调用下游任务执行服务时的协议，不属于 console/openapi 对外入口。
下游 HTTP client 由 `hz client` 生成：

```bash
hz client --idl idl/downstream/task_runner.thrift --module hz-server --client_dir biz/client --force_client
```

生成代码位于：

```text
biz/client/task_runner_service
biz/model/downstream/taskrunner
```

业务层不直接依赖生成 client，而是通过 `internal/taskrunner/client` 这个 adapter 包装后注入 `internal/task/service`。

这里刻意没有把 `--base_domain` 写进生成命令。下游地址来自配置文件：

```yaml
downstream:
  task_runner:
    endpoint: http://127.0.0.1:9000
```

这样不同环境只需要换配置，不需要重新生成代码。

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
biz/handler/console/task.StartTask -> internal/task/service -> internal/taskrunner/client -> downstream task runner

biz/handler/console/subtask.GetSubtask   -> internal/subtask/service -> main subtask repo      -> SQLite
biz/handler/console/subtask.ListSubtasks -> internal/subtask/service -> StarRocks subtask repo -> SQLite-simulated StarRocks
biz/handler/openapi/subtask              -> internal/subtask/service -> main subtask repo      -> SQLite
```

`console` 和 `openapi` 可以复用同一套内部业务能力，但各自拥有独立 IDL、路由、请求响应模型和 handler。

下游调用也按同样原则分层：

```text
idl/downstream/task_runner.thrift
  -> biz/model/downstream/taskrunner
  -> biz/client/task_runner_service
  -> internal/taskrunner/client
  -> internal/task/service.TaskRunner interface
```

各层职责：

```text
idl/downstream                 描述下游 HTTP 协议
biz/model/downstream           hz 生成的下游 request/response 模型
biz/client/task_runner_service hz 生成的 Hertz HTTP client
internal/taskrunner/client     手写 adapter，处理配置、超时、错误包装、模型转换
internal/task/service          只依赖 TaskRunner interface，不直接依赖 hz 生成代码
```

依赖方向是从外部实现流向内部接口：

```text
internal/task/service 定义需要什么能力
internal/taskrunner/client 实现这个能力
internal/app/container.go 把实现注入 service
```

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
  -> downstream.Init(cfg.Downstream)
      -> taskrunnerclient.New(...)
  -> app.NewContainer(dataSources, downstreamClients)
      -> tasksqlstore.New(dataSources.DB)
      -> taskrepo.New(taskSQL)
      -> taskservice.New(taskRepo, taskRunnerClient)
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

控制台开始任务接口会先查本地任务，再调用下游任务执行 client：

```text
POST /console/v1/tasks/:task_id/start?tenant_id=tenant-a
  -> biz/handler/console/task.StartTask
      -> task.StartTaskRequest
      -> taskservice.StartTaskInput
  -> internal/task/service.StartTask
      -> taskRepo.FindByTenantAndID
      -> taskRunner.StartTask
  -> internal/taskrunner/client
      -> taskrunner.StartTaskCommand
  -> hz generated client
  -> downstream task runner POST /tasks/start
```

`StartTaskInput` 是 service 层的用例入参：

```go
type StartTaskInput struct {
	TenantID string
	TaskID   int64
}
```

它和 IDL 生成的 `StartTaskRequest` 不是同一个模型。handler 负责把 HTTP 入参模型转换成 service 入参，service 再把这个 input 传给 `TaskRunner` 接口。真正调用下游时，`internal/taskrunner/client` 再把它转换成下游 IDL 生成的 `StartTaskCommand`。

这样做的好处是：

```text
console/openapi 的接口模型可以各自变化
service 的用例入参保持稳定
下游 task_runner 的协议变化被限制在 adapter 内
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

downstream:
  task_runner:
    endpoint: http://127.0.0.1:9000
    timeout_ms: 3000
```

`task_runner.endpoint` 是下游任务执行服务地址。`internal/taskrunner/client` 内部包装 hz 生成的 `TaskRunnerServiceClient`，调用下游 `POST /tasks/start`。

初始化时的实际过程：

```text
downstream.Init(cfg.Downstream)
  -> validate task_runner.endpoint
  -> validate task_runner.timeout_ms
  -> internal/taskrunner/client.New(...)
      -> biz/client/task_runner_service.NewTaskRunnerServiceClient(endpoint)
      -> return *internal/taskrunner/client.Client
  -> return downstream.Clients{TaskRunner: client}
```

然后在 container 中注入：

```text
app.NewContainer(ds, downstreamClients)
  -> taskservice.New(taskRepo, downstreamClients.TaskRunner)
```

所以 `task service` 拿到的是 `TaskRunner` 接口，不知道也不需要知道底层是 Hertz client、mock client，还是其他 RPC client。

HTTP 下游请求体：

```json
{
  "tenant_id": "tenant-a",
  "task_id": 1001
}
```

HTTP 下游响应体：

```json
{
  "accepted": true,
  "job_id": "job-http-1001",
  "message": "accepted by http runner"
}
```

启动时 `main.go` 会读取配置，再初始化容器、打开 SQLite、执行建表，并写入示例数据：

```text
internal/config      读取配置
internal/database    打开主库和 StarRocks 模拟库、建表、seed，并管理连接关闭
internal/downstream  初始化下游 client
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
curl -s -X POST 'http://127.0.0.1:8889/console/v1/tasks/1001/start?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/openapi/v1/tasks/1001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/console/v1/subtasks?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/console/v1/subtasks/5001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/openapi/v1/subtasks/5001?tenant_id=tenant-a'
```

console 响应包含 `tenant_id`、`owner`、`assignee` 等内部字段；openapi 响应暴露更少字段。
