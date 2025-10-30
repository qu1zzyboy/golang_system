package timeUtils

import (
	"fmt"
	"testing"
)

func TestGetNowDateStr(t *testing.T) {
	fmt.Println("GetNowDateStr:", GetNowDateStr())
	fmt.Println("GetNowTimeUnixMilli:", GetNowTimeUnixMilli())
	fmt.Println("GetNowMillisDateStrFast:", GetNowMillisDateStrFast())
}
