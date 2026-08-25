# 深潜布放放行台

深潜布放放行台为海洋观测任务提供任务登记、传感器配置核验、风险处置、安全复核、配置冻结和下潜放行凭据签发的一体化 JSON HTTP API。

## 构建

```bash
go build ./...
```

## 运行

```bash
go run ./cmd/deepdeploy -addr=127.0.0.1:19081
```

也可以通过 `PORT` 环境变量指定端口。使用 `-selfcheck` 可运行有界冒烟流程并自动退出：

```bash
go run ./cmd/deepdeploy -selfcheck -addr=127.0.0.1:19081
```

## 测试

```bash
go test ./...
```

主要接口包括任务和配置登记、配置核验、风险证据提交、安全复核及放行凭据签发。

任务创建支持 `Idempotency-Key`：同键同内容重试返回原任务，内容变化或重复任务编号返回 `409`。配置按 `revision` 保存完整历史并标记唯一活动版本；核验报告按 `contentHash` 缓存，任务查询可使用 `severity` 和 `openOnly` 过滤活动风险。已签发凭据可通过 `operation=revoke` 撤销，任务回到 `approved` 后可基于同一冻结配置重新签发，新旧凭据和审计事件均保留。
