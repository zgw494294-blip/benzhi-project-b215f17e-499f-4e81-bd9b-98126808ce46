package transport

// 路由由 Server.routes 注册；该文件保留路由清单，便于审计和文档生成。
var routeNames = []string{"POST /v1/tasks", "GET /v1/tasks/{taskID}", "POST /v1/tasks/{taskID}/configs", "POST /v1/tasks/{taskID}/configs/{configID}/validate", "POST /v1/tasks/{taskID}/risks/{riskID}/evidence", "POST /v1/tasks/{taskID}/review", "POST /v1/tasks/{taskID}/permit", "GET /v1/tasks/{taskID}/events", "GET /v1/tasks/{taskID}?severity=&openOnly="}
