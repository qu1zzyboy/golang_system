package defineTime

import (
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	now := time.Now()
	t.Log(now.Format(FormatHour)) // 输出格式化后的时间字符串
}
