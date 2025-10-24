package debugx

import (
	"path"
	"runtime"
	"strconv"
)

// 0	当前函数 GetCallerInfo()
// 1	当前函数调用者调用者函数
// 2	当前函数调用者的调用者的调用者函数
// 3	当前函数调用者的调用者的调用者的调用者函数

// GetCallerInfo  输出调用信息[文件名:行号]

func GetCaller(skip int) string {
	_, file, line, okCaller := runtime.Caller(skip)
	var callerInfo string
	if okCaller {
		callerInfo = "[" + path.Base(file) + ":" + strconv.Itoa(line) + "]"
	}
	return callerInfo
}
