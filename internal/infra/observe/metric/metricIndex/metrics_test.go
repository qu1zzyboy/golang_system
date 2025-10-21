package metricIndex

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/upbitBnServer/internal/infra/observe/metric/metricx"
)

func TestSystemMetrics(t *testing.T) {
	err := metricx.Init(metricx.Config{
		Enabled:     true,
		ServiceName: defineJson.QuantSystem,
		Port:        "2112",
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err = RegisterRuntimeMetrics()
	if err != nil {
		t.Fatalf("RegisterCallback failed: %v", err)
	}

	fmt.Println("访问 http://localhost:2112/metrics 查看系统指标")

	// 保持测试进程不退出,模拟运行中
	http.ListenAndServe(":2112", nil) // 持续运行暴露接口
}
