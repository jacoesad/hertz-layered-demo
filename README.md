# hertz-layered-demo

这是一个 CloudWeGo Hertz + `hz` 的分层架构示例项目。

项目重点不是实现复杂业务，而是演示：

- `hz new` / `hz update` 生成代码后，手写代码应该放在哪里
- `handler -> service -> repo -> sqlstore` 每层如何协作
- 配置、数据库、下游 client 这些变量如何从 `main` 显式传递
- 如何用接口做依赖倒置，避免业务层直接依赖数据库、Hertz client 或全局配置

当前示例包含 `task` 和 `subtask` 两个业务模块，并把 HTTP 入口拆成 `console` 和 `openapi` 两套接口。

## 整体结构

```text
idl/                         接口协议定义
biz/                         hz 生成的路由、handler 桩代码、model、client
internal/config              配置读取和校验
internal/database            数据库连接、建表、seed
internal/downstream          下游 client 初始化
internal/app                 应用容器，装配 service/repo/sqlstore
internal/task                task 领域代码
internal/subtask             subtask 领域代码
internal/taskrunner/client   下游 task runner 的手写 adapter
```

命名上按两个维度区分：

```text
downstream  表示依赖方向和业务角色，例如 idl/downstream、config.downstream、internal/downstream
client      表示调用实现和生成代码类型，例如 biz/client、internal/taskrunner/client
```

也就是说，`downstream` 说明“它是谁”，`client` 说明“我们怎么调用它”。

`biz` 目录按接口入口组织：

```text
biz/handler/console/task
biz/handler/console/subtask
biz/handler/openapi/task
biz/handler/openapi/subtask
```

`internal` 目录按业务领域组织：

```text
internal/task
internal/subtask
```

这样做的目的，是让 `console` 和 `openapi` 可以拥有不同的 IDL、路由、请求响应模型和 handler，但复用同一套内部业务能力。

## Main 启动顺序

`main.go` 是整个程序的装配入口。它负责读取配置、初始化外部资源、创建容器，最后启动 Hertz server。

```text
main.go
  -> config.Init("config/app.yaml")
      -> 读取 YAML
      -> 校验必填配置
      -> 返回 *config.Config

  -> database.Init(ctx, cfg)
      -> 打开主 SQLite: data/hz-server.db
      -> 打开模拟 StarRocks 的 SQLite: data/starrocks.db
      -> 建表
      -> 写入 demo 数据
      -> 返回 *database.DataSources

  -> downstream.Init(cfg.Downstream)
      -> 读取下游 task_runner endpoint/timeout
      -> 创建 internal/taskrunner/client
      -> 返回 *downstream.Clients

  -> app.NewContainer(dataSources, downstreamClients)
      -> 创建 sqlstore
      -> 创建 repo
      -> 创建 service
      -> 把依赖注入 service
      -> 返回 *app.Container

  -> app.Init(container)
      -> 保存 Default 容器，供 hz 生成的 handler 入口使用

  -> server.Default(server.WithHostPorts(cfg.Server.Addr))
  -> register(h)
  -> h.Spin()
```

对应代码的核心顺序是：

```go
cfg, err := config.Init(*configPath)

ds, err := database.Init(context.Background(), cfg)
defer ds.Close()

downstreamClients, err := downstream.Init(cfg.Downstream)

container, err := app.NewContainer(ds, downstreamClients)
app.Init(container)

h := server.Default(server.WithHostPorts(cfg.Server.Addr))
register(h)
h.Spin()
```

可以把 `main` 理解成唯一的组装层：它知道配置从哪里来，也知道哪些外部资源要创建；业务代码只接收自己真正需要的依赖。

`app.Default` 是为了适配 hz 生成的包级 handler 函数：router 调用的是固定函数入口，handler 再通过 `app.MustDefault()` 拿到已经装配好的 service。它保存的是装配结果，不是让 service/repo/client 自己去读取配置或创建依赖。

## 变量和依赖传递方式

项目刻意采用显式传递，而不是在业务代码里到处读全局变量。

推荐的依赖传递方向是：

```text
config/app.yaml
  -> config.Init
  -> cfg
  -> database.Init(cfg)
  -> downstream.Init(cfg.Downstream)
  -> app.NewContainer(ds, downstreamClients)
  -> service.New(repo, client)
  -> handler 调用 service
```

例如下游 task runner 的地址只在启动时读取一次：

```yaml
downstream:
  task_runner:
    endpoint: http://127.0.0.1:9000
    timeout_ms: 3000
```

然后按下面的路径传递：

```text
cfg.Downstream
  -> downstream.Init
  -> taskrunnerclient.New
  -> downstream.Clients{TaskRunner: client}
  -> app.NewContainer
  -> taskservice.New(taskRepo, clients.TaskRunner)
```

所以 `task service` 拿到的是 `TaskRunner` 接口，不知道底层是 Hertz 生成 client、mock client，还是别的 RPC client。

`internal/config` 包里保留了 `Get()` / `MustGet()`，主要是为了小项目或特殊场景兜底。当前主链路不依赖它们。这个项目更推荐：

```text
main 读取配置
main 创建依赖
main 把依赖注入 container/service
service/repo/client 不主动读取全局 config
```

这样依赖关系能从构造函数里看出来，测试时也更容易替换实现。

## 分层设计

以查询 task 为例，请求链路是：

```text
GET /console/v1/tasks/:task_id?tenant_id=tenant-a
  -> biz/router/console/task
  -> biz/handler/console/task.GetTask
  -> internal/task/service.GetTask
  -> internal/task/repo.FindByTenantAndID
  -> internal/task/sqlstore.SelectByTenantAndID
  -> SQLite tasks table
```

每层职责如下：

```text
handler   处理 HTTP 入参、调用 service、组装响应 DTO
service   表达业务用例和业务规则，只依赖接口
repo      把存储行模型转换成 domain 模型
sqlstore  执行 SQL，处理 database/sql 细节
domain    定义领域模型和领域错误
```

依赖方向保持单向：

```text
handler -> service -> repo -> sqlstore -> database/sql
```

其中 service 不直接依赖具体 repo 实现，而是定义自己需要的接口：

```go
type Repository interface {
	FindByTenantAndID(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error)
}

type TaskRunner interface {
	StartTask(ctx context.Context, input StartTaskInput) (*domain.StartTaskResult, error)
}
```

具体实现从外部注入：

```text
tasksqlstore.New(ds.DB)
  -> taskrepo.New(taskSQL)
  -> taskservice.New(taskRepo, clients.TaskRunner)
```

这就是这个项目里的依赖倒置：内部 service 定义需要什么能力，外部 repo/client 去实现这些能力，最后由 `app.NewContainer` 负责装配。

`subtask` 模块还演示了同一个 service 注入两个 repo：

```text
main subtask repo       -> data/hz-server.db
StarRocks subtask repo  -> data/starrocks.db
```

当前 `GetSubtask` 走主库，`ListSubtasks` 走模拟 StarRocks 的库：

```text
GetSubtask
  -> subtaskRepo.FindByTenantAndID
  -> main SQLite

ListSubtasks
  -> starRocksRepo.ListByTenant
  -> SQLite-simulated StarRocks
```

这只是为了演示读写来源可以在 service 层按用例选择；service 依赖的仍然是 repo 接口，而不是具体数据库。

## Handler 和 DTO 转换

`console` 和 `openapi` 有各自的 IDL model，所以 handler 不把 hz 生成的 request/response model 传进 service。

handler 的职责是做转换：

```text
IDL request model
  -> service input
  -> service method
  -> domain model
  -> IDL response model
```

DTO 转换放在各 handler 包的 `assembler.go` 中：

```text
biz/handler/console/task/assembler.go
biz/handler/openapi/task/assembler.go
biz/handler/console/subtask/assembler.go
biz/handler/openapi/subtask/assembler.go
```

assembler 只做模型转换，不查数据库、不读配置、不放业务规则。

这样 `console` 可以返回内部字段，比如 `tenant_id`、`owner`、`assignee`；`openapi` 可以暴露更少字段。两者的外部协议可以变化，但内部 service 不需要跟着频繁变化。

## 下游 Client 设计

`StartTask` 演示了如何调用下游任务执行服务。

请求链路是：

```text
POST /console/v1/tasks/:task_id/start?tenant_id=tenant-a
  -> biz/handler/console/task.StartTask
  -> internal/task/service.StartTask
      -> taskRepo.FindByTenantAndID
      -> taskRunner.StartTask
  -> internal/taskrunner/client
  -> biz/client/task_runner_service
  -> downstream task runner POST /tasks/start
```

这里有三类模型：

```text
console/task.StartTaskRequest        console 入口的 IDL 请求模型
taskservice.StartTaskInput           service 层用例入参
downstream/taskrunner.StartTaskCommand 下游 task_runner 的 IDL 请求模型
```

它们没有混用：

```text
handler 把 console request 转成 service input
service 使用 StartTaskInput 表达业务用例
taskrunner adapter 把 service input 转成下游 command
```

好处是：

```text
console/openapi 的接口模型可以各自变化
service 的用例入参保持稳定
下游 task_runner 的协议变化被限制在 adapter 内
```

`internal/taskrunner/client` 是手写 adapter。它包装 hz 生成的 client，负责超时、错误包装和模型转换：

```text
idl/downstream/task_runner.thrift
  -> biz/model/downstream/taskrunner
  -> biz/client/task_runner_service
  -> internal/taskrunner/client
  -> internal/task/service.TaskRunner interface
```

业务层不直接依赖 `biz/client/task_runner_service`，所以以后替换下游调用方式时，主要改 adapter 和装配代码。

## IDL 和生成代码

当前 IDL 按当前服务和下游协议拆分：

```text
idl/service/console/task.thrift       namespace go console.task
idl/service/console/subtask.thrift    namespace go console.subtask
idl/service/openapi/task.thrift       namespace go openapi.task
idl/service/openapi/subtask.thrift    namespace go openapi.subtask
idl/downstream/task_runner.thrift     namespace go downstream.taskrunner
```

日常修改 IDL 后，直接运行生成脚本：

```bash
script/hz_generate.sh
```

也可以只生成某一类代码：

```bash
script/hz_generate.sh server
script/hz_generate.sh client
```

脚本会自动发现 IDL 文件，不需要每新增一个 service 或下游 client 都改脚本：

```text
idl/service/**/*.thrift    -> hz update
idl/downstream/**/*.thrift -> hz client
```

等价于对扫描到的文件分别执行：

```bash
hz update --idl <server-idl>

hz client --idl <downstream-idl> --module hz-server --client_dir biz/client --force_client
```

这里没有把 `--base_domain` 写进生成命令。下游地址来自配置文件，这样不同环境只需要换配置，不需要重新生成代码。

首次创建 hz 项目骨架时使用的是：

```bash
hz new --module hz-server --idl idl/service/console/task.thrift
```

这个命令已经被收进脚本的 `init` 模式，但为了避免在已有项目里误跑，默认会拒绝执行：

```bash
script/hz_generate.sh init
HZ_FORCE_INIT=1 script/hz_generate.sh init
```

## 配置和运行

配置文件：

```text
config/app.yaml
```

当前配置项：

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

启动服务：

```bash
go run . -config config/app.yaml
```

本地数据库文件在：

```text
data/hz-server.db
data/starrocks.db
```

这两个文件由启动流程自动创建，已经加入 `.gitignore`。

## 示例接口

```bash
curl -s 'http://127.0.0.1:8889/console/v1/tasks/1001?tenant_id=tenant-a'
curl -s -X POST 'http://127.0.0.1:8889/console/v1/tasks/1001/start?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/openapi/v1/tasks/1001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/console/v1/subtasks?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/console/v1/subtasks/5001?tenant_id=tenant-a'
curl -s 'http://127.0.0.1:8889/openapi/v1/subtasks/5001?tenant_id=tenant-a'
```
