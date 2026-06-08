package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/config"
	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/models"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sheetsapi "google.golang.org/api/sheets/v4"
)

type Client struct {
	service       *sheetsapi.Service
	spreadsheetID string
	logger        *log.Logger
}

func NewClient(ctx context.Context, cfg *config.Config, logger *log.Logger) (*Client, error) {
	credentials, err := os.ReadFile(cfg.GoogleCredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать credentials.json: %w", err)
	}

	jwtConfig, err := google.JWTConfigFromJSON(credentials, sheetsapi.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать JWT-конфигурацию: %w", err)
	}

	httpClient := jwtConfig.Client(ctx)
	service, err := sheetsapi.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать клиент Google Sheets: %w", err)
	}

	client := &Client{
		service:       service,
		spreadsheetID: cfg.GoogleSpreadsheetID,
		logger:        logger,
	}

	if err := client.ensureHeaders(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) ensureHeaders(ctx context.Context) error {
	usersHeaders := []string{"TelegramID", "Username", "FullName", "Group"}
	if err := c.ensureSheetHeader(ctx, "Users", usersHeaders); err != nil {
		return err
	}

	attendanceHeaders := []string{"Date", "Time", "TelegramID", "Username", "FullName", "Group", "Status", "Reason"}
	return c.ensureSheetHeader(ctx, "Attendance", attendanceHeaders)
}

func (c *Client) ensureSheetHeader(ctx context.Context, sheet string, headers []string) error {
	rangeName := quoteSheetRange(sheet, fmt.Sprintf("A1:%s1", columnLetter(len(headers))))
	response, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, rangeName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("не удалось получить заголовок листа %s: %w", sheet, err)
	}

	if len(response.Values) == 0 || len(response.Values[0]) == 0 {
		valueRange := &sheetsapi.ValueRange{Values: [][]interface{}{toInterfaceSlice(headers)}}
		_, err := c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeName, valueRange).
			ValueInputOption("RAW").Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("не удалось установить заголовок листа %s: %w", sheet, err)
		}
	}

	return nil
}

func (c *Client) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	response, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, quoteSheetRange("Users", "A2:D")).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователей: %w", err)
	}

	for _, row := range response.Values {
		if len(row) < 4 {
			continue
		}

		rowID, err := strconv.ParseInt(fmt.Sprint(row[0]), 10, 64)
		if err != nil {
			continue
		}

		if rowID == telegramID {
			return &models.User{
				TelegramID: rowID,
				Username:   fmt.Sprint(row[1]),
				FullName:   fmt.Sprint(row[2]),
				Group:      fmt.Sprint(row[3]),
			}, nil
		}
	}

	return nil, nil
}

func (c *Client) CreateUser(ctx context.Context, user *models.User) error {
	values := []interface{}{strconv.FormatInt(user.TelegramID, 10), user.Username, user.FullName, user.Group}
	request := &sheetsapi.ValueRange{Values: [][]interface{}{values}}
	_, err := c.service.Spreadsheets.Values.Append(c.spreadsheetID, quoteSheetRange("Users", "A:D"), request).
		ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("не удалось сохранить пользователя: %w", err)
	}
	return nil
}

func (c *Client) CreateAttendance(ctx context.Context, attendance *models.Attendance) error {
	values := []interface{}{attendance.Date, attendance.Time, strconv.FormatInt(attendance.TelegramID, 10), attendance.Username, attendance.FullName, attendance.Group, attendance.Status, attendance.Reason}
	request := &sheetsapi.ValueRange{Values: [][]interface{}{values}}
	_, err := c.service.Spreadsheets.Values.Append(c.spreadsheetID, quoteSheetRange("Attendance", "A:H"), request).
		ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("не удалось сохранить заявку: %w", err)
	}
	return nil
}

func (c *Client) UpdateUserFullName(ctx context.Context, telegramID int64, fullName string) error {
	response, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, quoteSheetRange("Users", "A2:D")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("не удалось получить пользователей: %w", err)
	}

	for i, row := range response.Values {
		if len(row) < 1 {
			continue
		}

		rowID, err := strconv.ParseInt(fmt.Sprint(row[0]), 10, 64)
		if err != nil {
			continue
		}

		if rowID == telegramID {
			rangeName := quoteSheetRange("Users", fmt.Sprintf("C%d", i+2))
			valueRange := &sheetsapi.ValueRange{Values: [][]interface{}{{fullName}}}
			_, err := c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeName, valueRange).
				ValueInputOption("RAW").Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("не удалось обновить ФИО пользователя: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("пользователь не найден")
}

func toInterfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, value := range strings {
		result[i] = value
	}
	return result
}

func columnLetter(position int) string {
	if position <= 0 {
		return "A"
	}
	return fmt.Sprintf("%c", 'A'+position-1)
}

func quoteSheetRange(sheet, suffix string) string {
	sheet = strings.TrimSpace(sheet)
	if sheet == "" {
		return suffix
	}

	// Only quote if needed: spaces or special chars in sheet name.
	if strings.ContainsAny(sheet, " '!@#$%^&*()-+=,.;:/?[]{}~") {
		escaped := strings.ReplaceAll(sheet, "'", "''")
		return fmt.Sprintf("'%s'!%s", escaped, suffix)
	}

	return fmt.Sprintf("%s!%s", sheet, suffix)
}
