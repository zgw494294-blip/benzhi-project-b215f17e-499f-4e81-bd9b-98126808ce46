# BENZHI_README

基于 Go 实现的深潜布放放行台 HTTP API 项目，一款后端服务，深潜布放放行台已实现任务登记、传感器配置版本管理、自动核验、风险证据处置、安全复核、配置冻结和不可变下潜放行凭据签发，并通过可配置回环地址提供 JSON HTTP API 与有界 selfcheck。

## 项目说明
- 项目：benzhi-project-b215f17e-499f-4e81-bd9b-98126808ce46
- 项目用途：深潜布放放行台已实现任务登记、传感器配置版本管理、自动核验、风险证据处置、安全复核、配置冻结和不可变下潜放行凭据签发，并通过可配置回环地址提供 JSON HTTP API 与有界 selfcheck。
- Go 工具链：`golang:1.22`
- 前端工具链：无

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/deepdeploy -selfcheck -addr=127.0.0.1:19081
cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-project-b215f17e-499f-4e81-bd9b-98126808ce46-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-project-b215f17e-499f-4e81-bd9b-98126808ce46-arm64 linux/arm64
docker run -it benzhi-project-b215f17e-499f-4e81-bd9b-98126808ce46-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/deepdeploy -selfcheck -addr=127.0.0.1:19081`
