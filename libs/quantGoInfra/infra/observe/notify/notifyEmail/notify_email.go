package notifyEmail

import (
	"github.com/hhh500/quantGoInfra/pkg/utils/jsonUtils"
	"gopkg.in/gomail.v2"
)

type NotifyEmail struct {
}

func NewNotifyTg() *NotifyEmail {
	return &NotifyEmail{}
}

func (s *NotifyEmail) SendImportantErrorMsg(payload map[string]string) error {
	return SendMailJsonByGomail("", "重要错误通知", payload)
}

func (s *NotifyEmail) SendNormalErrorMsg(payload map[string]string) error {
	return SendMailJsonByGomail("", "重要错误通知", payload)
}

func (s *NotifyEmail) SendReminderMsg(payload map[string]string) error {
	return SendMailJsonByGomail("", "重要错误通知", payload)
}

type MyEmailMsg struct {
	Title  string `json:"title"`
	Header string `json:"header"`
	Body   string `json:"body"`
	To     string `json:"to"`
}

func SendMailJsonByGomail(title, header string, payload map[string]string) error {
	jsonData, err := jsonUtils.MarshalStructToString(payload)
	if err != nil {
		return err
	}
	return Send163MailByGomail(title, header, jsonData)
}

func Send163MailByGomail(title, header, text string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "2522558991@qq.com", title)
	m.SetHeader("To", "2282915646@qq.com")
	m.SetHeader("Subject", header)
	m.SetBody("text/html", text)
	d := gomail.NewDialer("smtp.qq.com", 465, "2522558991@qq.com", "xogoebtwrcfuecbe")
	return d.DialAndSend(m)
}
