package notify

type Notify interface {
	SendImportantErrorMsg(payload map[string]string) error // 发送重要错误消息
	SendNormalErrorMsg(payload map[string]string) error    // 发送普通错误消息
	SendReminderMsg(payload map[string]string) error       // 发送提醒消息
}

var impl Notify

func setNotify(impl_ Notify) {
	impl = impl_
}

func GetNotify() Notify {
	return impl
}
