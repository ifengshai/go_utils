// 测试用日志写入器 - 每秒写入 INFO / ERROR / NOTICE 三条日志
// 运行方式：go test ./log_monitoring/testdata/ -v -run TestWriteLog -timeout 0
package testdata

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteLog(t *testing.T) {
	logPath := filepath.Join("test.log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("cannot open log file: %v", err)
	}
	defer f.Close()

	t.Logf("writing log to: %s  (press Ctrl+C to stop)", logPath)

	for i := 1; ; i++ {
		now := time.Now().Format("2006-01-02 15:04:05")

		lines := []string{
			fmt.Sprintf("[%s] [INFO]   seq=%d  system running normally, all OK\n", now, i),
			fmt.Sprintf("[%s] [ERROR]  seq=%d  simulated error: connection timeout (code=500)\n", now, i),
			fmt.Sprintf("[%s] [NOTICE] seq=%d  disk usage reached 80%%\n", now, i),
		}

		for _, line := range lines {
			if _, err := f.WriteString(line); err != nil {
				t.Errorf("write failed: %v", err)
				return
			}
		}

		// 立即刷盘，确保监控工具能实时读到
		if err := f.Sync(); err != nil {
			t.Errorf("sync failed: %v", err)
			return
		}

		time.Sleep(time.Second)
	}
}
