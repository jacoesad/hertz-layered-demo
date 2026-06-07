namespace go console.subtask

struct GetSubtaskRequest {
    1: required string tenant_id (api.query="tenant_id")
    2: required i64 subtask_id (api.path="subtask_id")
}

struct ConsoleSubtaskInfo {
    1: i64 id (api.body="id")
    2: string tenant_id (api.body="tenant_id")
    3: i64 task_id (api.body="task_id")
    4: string title (api.body="title")
    5: string status (api.body="status")
    6: string assignee (api.body="assignee")
}

struct GetSubtaskResponse {
    1: i32 code (api.body="code")
    2: string message (api.body="message")
    3: optional ConsoleSubtaskInfo data (api.body="data")
}

struct ListSubtasksRequest {
    1: required string tenant_id (api.query="tenant_id")
    2: optional i64 task_id (api.query="task_id")
    3: optional string subtask_type (api.query="subtask_type")
}

struct ListSubtasksResponse {
    1: i32 code (api.body="code")
    2: string message (api.body="message")
    3: list<ConsoleSubtaskInfo> data (api.body="data")
}

service ConsoleSubtaskService {
    GetSubtaskResponse GetSubtask(1: GetSubtaskRequest req) (api.get="/console/v1/subtasks/:subtask_id")
    ListSubtasksResponse ListSubtasks(1: ListSubtasksRequest req) (api.get="/console/v1/subtasks")
}
