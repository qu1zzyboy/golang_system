package notifyTg

// import (
// 	"QuantGo/pkg/utils/jsonUtils"
// 	tgbotApi "github.com/go-telegram-bot-api/telegram-bot-api"
// )

// type NotifyTg struct {
// 	bot *tgbotApi.BotAPI
// }

// func NewNotifyTg() (*NotifyTg, error) {
// 	bot, err := tgbotApi.NewBotAPI("7164045954:AAHLGtw9YIe6EnQmaCJxObF-Imxr5MIxjr4")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &NotifyTg{bot: bot}, nil
// }

// func (s *NotifyTg) SendImportantErrorMsg(payload map[string]string) error {
// 	jsonData, err := jsonUtils.MarshalStructToString(payload)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = s.bot.Send(tgbotApi.NewMessage(-4934397094, jsonData))
// 	return err
// }

// func (s *NotifyTg) SendNormalErrorMsg(payload map[string]string) error {
// 	jsonData, err := jsonUtils.MarshalStructToString(payload)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = s.bot.Send(tgbotApi.NewMessage(-4778765274, jsonData))
// 	return err
// }

// func (s *NotifyTg) SendReminderMsg(payload map[string]string) error {
// 	jsonData, err := jsonUtils.MarshalStructToString(payload)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = s.bot.Send(tgbotApi.NewMessage(6769909558, jsonData))
// 	return err
// }
