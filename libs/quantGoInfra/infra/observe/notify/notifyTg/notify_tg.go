package notifyTg

import (
	tgbotApi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/quantGoInfra/pkg/utils/jsonUtils"
)

var serviceSingleton = singleton.NewSingleton(func() *NotifyTg {
	return NewNotifyTg()
})

func GetTg() *NotifyTg {
	return serviceSingleton.Get()
}

type NotifyTg struct {
	bot      *tgbotApi.BotAPI
	botUpBit *tgbotApi.BotAPI
}

// https://api.telegram.org/bot<你的TOKEN>/getUpdates

func NewNotifyTg() *NotifyTg {
	bot, err := tgbotApi.NewBotAPI("7164045954:AAHLGtw9YIe6EnQmaCJxObF-Imxr5MIxjr4")
	if err != nil {
		panic(err)
	}
	botUpBit, err := tgbotApi.NewBotAPI("8447121214:AAGkfJ7tsA16ODDXoXcuwSteVAgShPQaR1c")
	if err != nil {
		panic(err)
	}
	return &NotifyTg{bot: bot, botUpBit: botUpBit}
}

func (s *NotifyTg) SendToUpBitMsg(payload map[string]string) error {
	jsonData, err := jsonUtils.MarshalStructToString(payload)
	if err != nil {
		return err
	}
	_, err = s.botUpBit.Send(tgbotApi.NewMessage(-4771082539, jsonData))
	return err
}

func (s *NotifyTg) SendToUpBitStrMsg(msg string) error {
	_, err := s.botUpBit.Send(tgbotApi.NewMessage(-4771082539, msg))
	return err
}

func (s *NotifyTg) SendImportantErrorMsg(payload map[string]string) error {
	jsonData, err := jsonUtils.MarshalStructToString(payload)
	if err != nil {
		return err
	}
	_, err = s.bot.Send(tgbotApi.NewMessage(-4934397094, jsonData))
	return err
}

func (s *NotifyTg) SendNormalErrorMsg(payload map[string]string) error {
	jsonData, err := jsonUtils.MarshalStructToString(payload)
	if err != nil {
		return err
	}
	_, err = s.bot.Send(tgbotApi.NewMessage(-4778765274, jsonData))
	return err
}

func (s *NotifyTg) SendReminderMsg(payload map[string]string) error {
	jsonData, err := jsonUtils.MarshalStructToString(payload)
	if err != nil {
		return err
	}
	_, err = s.bot.Send(tgbotApi.NewMessage(6769909558, jsonData))
	return err
}
