package transport

import (
	"bytes"
	"deepdeploy/internal/domain"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

func (s *Server) SelfCheck() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	go s.http.Serve(ln)
	time.Sleep(30 * time.Millisecond)
	base := "http://" + s.addr
	post := func(path string, v any, headers map[string]string) (map[string]any, error) {
		b, _ := json.Marshal(v)
		req, _ := http.NewRequest("POST", base+path, bytes.NewReader(b))
		for k, val := range headers {
			req.Header.Set(k, val)
		}
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			return nil, e
		}
		defer resp.Body.Close()
		var out map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&out)
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("%s: %d", path, resp.StatusCode)
		}
		return out, nil
	}
	t, err := post("/v1/tasks", map[string]string{"id": "self-task", "missionName": "自检任务", "seaArea": "A", "deploymentWindow": "2026-08-25T00:00:00Z/2026-08-26T00:00:00Z", "owner": "工程师"}, nil)
	if err != nil {
		return err
	}
	_ = t
	// 使用有效配置直接完成核验、复核和放行链路。
	conf := map[string]any{"sensors": []domain.Sensor{{ID: "s1", Model: "M", Bus: "bus1", SampleRate: 10}}, "firmwareSet": map[string]string{"s1": "1.0"}, "mountingParameters": map[string]float64{"depth_m": 100}, "environmentLimits": map[string]float64{"max_pressure_bar": 100, "max_depth_m": 1000}, "submittedBy": "工程师"}
	c, err := post("/v1/tasks/self-task/configs", conf, map[string]string{"X-Expected-Version": "1", "Idempotency-Key": "self-config"})
	if err != nil {
		return err
	}
	_ = c
	// 获取配置 ID 以便继续流程。
	resp, err := http.Get(base + "/v1/tasks/self-task")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var task domain.Task
	if err = json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return err
	}
	if _, err = post("/v1/tasks/self-task/configs/"+task.ActiveConfigID+"/validate", map[string]any{}, map[string]string{"X-Expected-Version": "2"}); err != nil {
		return err
	}
	if _, err = post("/v1/tasks/self-task/review", map[string]string{"decision": "approve", "reviewer": "安全负责人"}, map[string]string{"X-Expected-Version": "3"}); err != nil {
		return err
	}
	if _, err = post("/v1/tasks/self-task/permit", map[string]string{"issuer": "安全负责人"}, map[string]string{"X-Expected-Version": "4"}); err != nil {
		return err
	}
	return nil
}
