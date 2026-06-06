namespace go downstream.taskrunner

struct StartTaskCommand {
    1: required string tenant_id (api.body="tenant_id")
    2: required i64 task_id (api.body="task_id")
    3: optional string operator (api.body="operator")
}

struct StartTaskResult {
    1: required bool accepted (api.body="accepted")
    2: optional string job_id (api.body="job_id")
    3: optional string message (api.body="message")
}

service TaskRunnerService {
    StartTaskResult StartTask(1: StartTaskCommand req) (api.post="/tasks/start")
}
