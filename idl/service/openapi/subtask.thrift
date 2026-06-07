namespace go openapi.subtask

struct QuerySubtaskRequest {
    1: required string tenant_id (api.query="tenant_id")
    2: required i64 subtask_id (api.path="subtask_id")
}

struct OpenAPISubtaskInfo {
    1: i64 id (api.body="id")
    2: i64 task_id (api.body="task_id")
    3: string title (api.body="title")
    4: string status (api.body="status")
}

struct QuerySubtaskResponse {
    1: i32 code (api.body="code")
    2: string message (api.body="message")
    3: optional OpenAPISubtaskInfo data (api.body="data")
}

service OpenAPISubtaskService {
    QuerySubtaskResponse QuerySubtask(1: QuerySubtaskRequest req) (api.get="/openapi/v1/subtasks/:subtask_id")
}
