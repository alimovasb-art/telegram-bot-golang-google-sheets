package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	StatusNotComing = "not_coming"
	StatusLate      = "late"
	StatusSick      = "sick"
	StatusOnline    = "online"
)

var StatusLabels = map[string]string{
	StatusNotComing: "❌ Не приду",
	StatusLate:      "⏰ Опоздаю",
	StatusSick:      "🏥 Болею",
	StatusOnline:    "💻 Зайду на онлайн урок",
}

func MainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(StatusLabels[StatusNotComing], StatusNotComing),
			tgbotapi.NewInlineKeyboardButtonData(StatusLabels[StatusLate], StatusLate),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(StatusLabels[StatusSick], StatusSick),
			tgbotapi.NewInlineKeyboardButtonData(StatusLabels[StatusOnline], StatusOnline),
		),
	)
}
