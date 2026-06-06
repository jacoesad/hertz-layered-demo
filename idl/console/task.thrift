namespace go console.task

struct GetTaskRequest {
    1: required string tenant_id (api.query="tenant_id")
    2: required i64 task_id (api.path="task_id")
}

struct ConsoleTaskInfo {
    1: i64 id (api.body="id")
    2: string tenant_id (api.body="tenant_id")
    3: string title (api.body="title")
    4: string status (api.body="status")
    5: string owner (api.body="owner")
}

struct GetTaskResponse {
    1: i32 code (api.body="code")
    2: string message (api.body="message")
    3: optional ConsoleTaskInfo data (api.body="data")
}

service ConsoleTaskService {
    GetTaskResponse GetTask(1: GetTaskRequest req) (api.get="/console/v1/tasks/:task_id")
}
