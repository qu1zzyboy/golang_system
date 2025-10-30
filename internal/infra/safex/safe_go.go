package safex

import (
	"runtime/debug"

	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/observe/notify"
)

func SafeGo(protectId string, fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				staticLog.LogPanic.Info()
				staticLog.LogPanic.Infof("panic: %v\n%s", err, debug.Stack())
				notify.GetNotify().SendImportantErrorMsg(map[string]string{defineJson.Msg: "panic捕获", "protectId": conf.ServerName + "_" + protectId})
			}
		}()
		fn()
	}()
}

func SafeGoWrap(protectId string, fn func()) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				staticLog.LogPanic.Info()
				staticLog.LogPanic.Infof("panic: %v\n%s", err, debug.Stack())
				notify.GetNotify().SendImportantErrorMsg(map[string]string{defineJson.Msg: "panic捕获", "protectId": conf.ServerName + "_" + protectId})
			}
		}()
		fn()
	}
}
