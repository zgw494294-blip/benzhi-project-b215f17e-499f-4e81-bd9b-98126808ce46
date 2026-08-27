package malformedsnapshot

import (
	"os"
	"path/filepath"
	"testing"

	"deepdeploy/internal/persistence"
)

func TestMalformedSnapshotRejected(t *testing.T) {
	cases := map[string]string{
		"config": `{"cfg-1":null}`,
		"risk":   `{"risk-1":null}`,
	}
	for name, members := range cases {
		members := members
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			configs, risks := `{}`, members
			if name == "config" {
				configs, risks = members, `{}`
			}
			snapshot := []byte(`{"id":"task-1","missionName":"任务","seaArea":"海域","deploymentWindow":"窗口","owner":"负责人","status":"draft","version":1,"configs":` + configs + `,"risks":` + risks + `}`)
			if err := os.WriteFile(filepath.Join(dir, "task-1.json"), snapshot, 0600); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("malformed snapshot panic: %v", recovered)
				}
			}()
			if _, err := persistence.NewFileStore(dir); err == nil {
				t.Fatal("malformed snapshot accepted")
			}
		})
	}
}
