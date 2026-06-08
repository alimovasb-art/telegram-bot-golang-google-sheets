package app

import (
	"context"
	"log"
	"os"

	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/config"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/fsm"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/handlers"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/sheets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Run(ctx context.Context) error {
	logger := log.New(os.Stdout, "[attendance-bot] ", log.LstdFlags|log.Lmsgprefix)

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return err
	}
	bot.Debug = false

	logger.Printf("Запуск Telegram Attendance Bot в аккаунте @%s", bot.Self.UserName)

	sheetsClient, err := sheets.NewClient(ctx, cfg, logger)
	if err != nil {
		return err
	}

	handler := handlers.NewBotHandler(bot, sheetsClient, fsm.NewStore(), logger)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		handler.HandleUpdate(ctx, update)
	}

	logger.Println("Update channel закрыт, остановка бота")
	return nil
}
