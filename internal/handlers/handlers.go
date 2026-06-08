package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/fsm"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/models"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/sheets"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot    *tgbotapi.BotAPI
	sheets *sheets.Client
	fsm    *fsm.Store
	logger *log.Logger
}

func NewBotHandler(bot *tgbotapi.BotAPI, sheetsClient *sheets.Client, fsmStore *fsm.Store, logger *log.Logger) *BotHandler {
	return &BotHandler{
		bot:    bot,
		sheets: sheetsClient,
		fsm:    fsmStore,
		logger: logger,
	}
}

func (h *BotHandler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message != nil {
		h.handleMessage(ctx, update.Message)
		return
	}

	if update.CallbackQuery != nil {
		h.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}
}

func (h *BotHandler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	if message.IsCommand() {
		h.handleCommand(ctx, message)
		return
	}

	state := h.fsm.GetState(message.From.ID)
	switch state {
	case models.StateWaitingFullName:
		h.processFullName(ctx, message)
	case models.StateWaitingFullNameUpdate:
		h.processFullNameUpdate(ctx, message)
	case models.StateWaitingGroup:
		h.processGroup(ctx, message)
	case models.StateWaitingReason:
		h.processReason(ctx, message)
	default:
		h.replyText(message.Chat.ID, "Используйте /start, чтобы начать работу, или выберите действие в меню.")
	}
}

func (h *BotHandler) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		h.handleStart(ctx, message)
	case "profile":
		h.handleProfile(ctx, message)
	case "change_name":
		h.handleChangeName(ctx, message)
	case "help":
		h.handleHelp(ctx, message)
	default:
		h.replyText(message.Chat.ID, "Неизвестная команда. Доступные команды: /start, /profile, /change_name, /help")
	}
}

func (h *BotHandler) handleHelp(ctx context.Context, message *tgbotapi.Message) {
	helpText := "Доступные команды:\n" +
		"/start - регистрация и начало работы\n" +
		"/profile - показывает ваш профиль\n" +
		"/change_name - изменить ФИО, если оно указано неверно\n" +
		"/help - показать это сообщение"
	h.replyText(message.Chat.ID, helpText)
}

func (h *BotHandler) handleStart(ctx context.Context, message *tgbotapi.Message) {
	user, err := h.sheets.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		h.replyError(message.Chat.ID, "Ошибка доступа к Google Sheets. Попробуйте позже.")
		h.logger.Printf("GetUserByTelegramID error: %v", err)
		return
	}

	if user != nil {
		h.replyText(message.Chat.ID, fmt.Sprintf("Вы уже зарегистрированы как %s.\n\nДобро пожаловать обратно!", user.FullName))
		h.sendMainMenu(message.Chat.ID)
		return
	}

	h.fsm.SetState(message.From.ID, models.StateWaitingFullName)
	h.fsm.SetRegistrationData(message.From.ID, &models.RegistrationData{})
	h.replyText(message.Chat.ID, "Добро пожаловать! Пожалуйста, введите ваше ФИО для регистрации:")
}

func (h *BotHandler) handleProfile(ctx context.Context, message *tgbotapi.Message) {
	user, err := h.sheets.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		h.replyError(message.Chat.ID, "Ошибка доступа к Google Sheets. Попробуйте позже.")
		h.logger.Printf("GetUserByTelegramID error: %v", err)
		return
	}

	if user == nil {
		h.replyText(message.Chat.ID, "Вы не зарегистрированы. Используйте /start для регистрации.")
		return
	}

	username := user.Username
	if username == "" {
		username = "-"
	}

	text := fmt.Sprintf("Ваш профиль:\nФИО: %s\nГруппа: %s\nUsername: %s\nTelegram ID: %d\n\nЕсли ФИО указано неверно, используйте команду /change_name.", user.FullName, user.Group, username, user.TelegramID)
	h.replyText(message.Chat.ID, text)
}

func (h *BotHandler) handleChangeName(ctx context.Context, message *tgbotapi.Message) {
	user, err := h.sheets.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		h.replyError(message.Chat.ID, "Ошибка доступа к Google Sheets. Попробуйте позже.")
		h.logger.Printf("GetUserByTelegramID error: %v", err)
		return
	}

	if user == nil {
		h.replyText(message.Chat.ID, "Вы не зарегистрированы. Используйте /start для регистрации.")
		return
	}

	h.fsm.SetState(message.From.ID, models.StateWaitingFullNameUpdate)
	h.replyText(message.Chat.ID, fmt.Sprintf("Ваше текущее ФИО: %s\nВведите правильное ФИО:", user.FullName))
}

func (h *BotHandler) processFullName(ctx context.Context, message *tgbotapi.Message) {
	fullName := strings.TrimSpace(message.Text)
	if fullName == "" {
		h.replyText(message.Chat.ID, "ФИО не может быть пустым. Пожалуйста, введите ФИО:")
		return
	}

	h.fsm.SetRegistrationData(message.From.ID, &models.RegistrationData{FullName: fullName})
	h.fsm.SetState(message.From.ID, models.StateWaitingGroup)
	h.replyText(message.Chat.ID, "Спасибо. Введите вашу группу:")
}

func (h *BotHandler) processFullNameUpdate(ctx context.Context, message *tgbotapi.Message) {
	fullName := strings.TrimSpace(message.Text)
	if fullName == "" {
		h.replyText(message.Chat.ID, "ФИО не может быть пустым. Пожалуйста, введите ФИО:")
		return
	}

	if err := h.sheets.UpdateUserFullName(ctx, message.From.ID, fullName); err != nil {
		h.replyError(message.Chat.ID, "Не удалось обновить ФИО. Попробуйте позже.")
		h.logger.Printf("UpdateUserFullName error: %v", err)
		return
	}

	h.fsm.Clear(message.From.ID)
	h.replyText(message.Chat.ID, fmt.Sprintf("Ваше ФИО успешно обновлено на: %s", fullName))
	h.sendMainMenu(message.Chat.ID)
}

func (h *BotHandler) processGroup(ctx context.Context, message *tgbotapi.Message) {
	group := strings.TrimSpace(message.Text)
	if group == "" {
		h.replyText(message.Chat.ID, "Группа не может быть пустой. Пожалуйста, введите группу:")
		return
	}

	registration := h.fsm.GetRegistrationData(message.From.ID)
	if registration == nil || registration.FullName == "" {
		h.replyText(message.Chat.ID, "Произошла ошибка регистрации. Начните заново командой /start.")
		h.fsm.Clear(message.From.ID)
		return
	}

	registration.Group = group
	user := &models.User{
		TelegramID: message.From.ID,
		Username:   message.From.UserName,
		FullName:   registration.FullName,
		Group:      registration.Group,
	}

	if err := h.sheets.CreateUser(ctx, user); err != nil {
		h.replyError(message.Chat.ID, "Не удалось сохранить пользователя. Попробуйте позже.")
		h.logger.Printf("CreateUser error: %v", err)
		return
	}

	h.fsm.Clear(message.From.ID)
	h.replyText(message.Chat.ID, "Регистрация завершена. Ваша информация сохранена.")
	h.sendMainMenu(message.Chat.ID)
}

func (h *BotHandler) processReason(ctx context.Context, message *tgbotapi.Message) {
	reason := strings.TrimSpace(message.Text)
	if reason == "" {
		h.replyText(message.Chat.ID, "Причина не может быть пустой. Пожалуйста, опишите причину:")
		return
	}

	registration := h.fsm.GetRegistrationData(message.From.ID)
	if registration == nil || registration.PendingStatus == "" {
		h.replyText(message.Chat.ID, "Выберите статус через главное меню, прежде чем отправлять заявку.")
		h.fsm.Clear(message.From.ID)
		return
	}

	user, err := h.sheets.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		h.replyError(message.Chat.ID, "Ошибка доступа к Google Sheets. Попробуйте позже.")
		h.logger.Printf("GetUserByTelegramID error: %v", err)
		return
	}
	if user == nil {
		h.replyText(message.Chat.ID, "Вы не зарегистрированы. Используйте /start для регистрации.")
		h.fsm.Clear(message.From.ID)
		return
	}

	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		h.logger.Printf("LoadLocation Asia/Tashkent error: %v; falling back to +05:00", err)
		loc = time.FixedZone("Tashkent", 5*3600)
	}
	timestamp := time.Now().In(loc)
	statusLabel := telegram.StatusLabels[registration.PendingStatus]
	attendance := &models.Attendance{
		Date:       timestamp.Format("2006-01-02"),
		Time:       timestamp.Format("15:04:05"),
		TelegramID: user.TelegramID,
		Username:   user.Username,
		FullName:   user.FullName,
		Group:      user.Group,
		Status:     statusLabel,
		Reason:     reason,
	}

	if err := h.sheets.CreateAttendance(ctx, attendance); err != nil {
		h.replyError(message.Chat.ID, "Не удалось сохранить заявку. Попробуйте позже.")
		h.logger.Printf("CreateAttendance error: %v", err)
		return
	}

	h.fsm.Clear(message.From.ID)
	h.replyText(message.Chat.ID, "Ваша заявка успешно сохранена. Спасибо! Всего доброго.")
}

func (h *BotHandler) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	if query.Message == nil {
		return
	}

	statusKey := query.Data
	statusLabel, ok := telegram.StatusLabels[statusKey]
	if !ok {
		h.answerCallback(query, "Неизвестный статус")
		return
	}

	user, err := h.sheets.GetUserByTelegramID(ctx, query.From.ID)
	if err != nil {
		h.replyError(query.Message.Chat.ID, "Ошибка доступа к Google Sheets. Попробуйте позже.")
		h.logger.Printf("GetUserByTelegramID error: %v", err)
		h.answerCallback(query, "Ошибка при обработке запроса")
		return
	}

	if user == nil {
		h.answerCallback(query, "Сначала зарегистрируйтесь через /start")
		h.replyText(query.Message.Chat.ID, "Вам нужно зарегистрироваться через /start перед подачей заявки.")
		return
	}

	registration := h.fsm.GetRegistrationData(query.From.ID)
	if registration == nil {
		registration = &models.RegistrationData{}
	}
	registration.PendingStatus = statusKey
	h.fsm.SetRegistrationData(query.From.ID, registration)
	h.fsm.SetState(query.From.ID, models.StateWaitingReason)

	h.answerCallback(query, fmt.Sprintf("Вы выбрали %s", statusLabel))
	h.replyText(query.Message.Chat.ID, "Пожалуйста, укажите причину вашей заявки:")
}

func (h *BotHandler) sendMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Главное меню. Выберите статус:")
	msg.ReplyMarkup = telegram.MainMenuKeyboard()
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Printf("sendMainMenu error: %v", err)
	}
}

func (h *BotHandler) replyText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Printf("replyText error: %v", err)
	}
}

func (h *BotHandler) replyError(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Printf("replyError send failed: %v", err)
	}
}

func (h *BotHandler) answerCallback(query *tgbotapi.CallbackQuery, text string) {
	callback := tgbotapi.NewCallback(query.ID, text)
	if _, err := h.bot.Request(callback); err != nil {
		h.logger.Printf("answerCallback error: %v", err)
	}
}
