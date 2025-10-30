package defineTime

const (
	FormatCmc     = "2006-01-02T15:04:05.000Z"
	FormatSec     = "2006-01-02 15:04:05"
	FormatMillSec = "2006-01-02 15:04:05.000"
	FormatDate    = "2006-01-02"
	FormatHour    = "2006-01-02-15"
)

const (
	DayBeginStr    = "1 0 8 * * *"                  // 每天的 08:00:01 执行任务
	HourEndStr     = "59 59 * * * *"                // 每小时的最后一秒执行任务
	HourBegin6MStr = "01 06 * * * *"                // 每小时的最后一秒执行任务
	Min10EndStr    = "59 09,19,29,39,49,59 * * * *" // 每10分钟结束执行时间
)

const (
	Sec10EndStr_1 = "09 * * * * *" // crontab每10秒结束执行时间1
	Sec10EndStr_2 = "19 * * * * *" // crontab每10秒结束执行时间2
	Sec10EndStr_3 = "29 * * * * *" // crontab每10秒结束执行时间3
	Sec10EndStr_4 = "39 * * * * *" // crontab每10秒结束执行时间4
	Sec10EndStr_5 = "49 * * * * *" // crontab每10秒结束执行时间5
	Sec10EndStr_6 = "59 * * * * *" // crontab每10秒结束执行时间6

	Sec10PreEndStr_1 = "08 * * * * *" // crontab每10秒结束执行时间1
	Sec10PreEndStr_2 = "18 * * * * *" // crontab每10秒结束执行时间2
	Sec10PreEndStr_3 = "28 * * * * *" // crontab每10秒结束执行时间3
	Sec10PreEndStr_4 = "38 * * * * *" // crontab每10秒结束执行时间4
	Sec10PreEndStr_5 = "48 * * * * *" // crontab每10秒结束执行时间5
	Sec10PreEndStr_6 = "58 * * * * *" // crontab每10秒结束执行时间6

	MinEndStr_59 = "59 * * * * *" // crontab每分钟结束执行时间
	MinEndStr_58 = "58 * * * * *" // crontab每分钟结束执行时间
	MinEndStr_49 = "49 * * * * *" // crontab每分钟结束执行时间

	MinEndStr_0 = "59 0/10 * * * *" // crontab 第0分钟结束执行时间
	MinEndStr_1 = "59 1/10 * * * *" // crontab 第1分钟结束执行时间
	MinEndStr_2 = "59 2/10 * * * *" // crontab 第2分钟结束执行时间
	MinEndStr_3 = "59 3/10 * * * *" // crontab 第3分钟结束执行时间
	MinEndStr_4 = "59 4/10 * * * *" // crontab 第4分钟结束执行时间
	MinEndStr_5 = "59 5/10 * * * *" // crontab 第5分钟结束执行时间
	MinEndStr_6 = "59 6/10 * * * *" // crontab 第6分钟结束执行时间
	MinEndStr_7 = "59 7/10 * * * *" // crontab 第7分钟结束执行时间
	MinEndStr_8 = "59 8/10 * * * *" // crontab 第8分钟结束执行时间
	MinEndStr_9 = "59 9/10 * * * *" // crontab 第9分钟结束执行时间
)
