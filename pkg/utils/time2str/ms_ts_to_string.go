package time2str

import (
	"strconv"
	"time"
	"upbitBnServer/internal/infra/systemx"
)

func GetNowTimeStampMicroSlice16() systemx.WsId16B {
	us := time.Now().UnixMicro()
	return systemx.WsId16B(strconv.FormatInt(us, 10))
}
