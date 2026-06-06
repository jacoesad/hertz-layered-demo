namespace go openapi.task

struct QueryTaskRequest {
    1: required string tenant_id (api.query="tenant_id")
    2: required i64 task_id (api.path="task_id")
}

struct OpenAPITaskInfo {
    1: i64 id (api.body="id")
    2: string title (api.body="title")
    3: string status (api.body="status")
}

struct QueryTaskResponse {
    1: i32 code (api.body="code")
    2: string message (api.body="message")
    3: optional OpenAPITaskInfo data (api.body="data")
}

service OpenAPITaskService {
    QueryTaskResponse QueryTask(1: QueryTaskRequest req) (api.get="/openapi/v1/tasks/:task_id")
}
