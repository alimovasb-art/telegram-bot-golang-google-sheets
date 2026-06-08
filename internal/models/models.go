package models

type State string

const (
	StateIdle                  State = "idle"
	StateWaitingFullName       State = "waiting_full_name"
	StateWaitingGroup          State = "waiting_group"
	StateWaitingReason         State = "waiting_reason"
	StateWaitingFullNameUpdate State = "waiting_full_name_update"
)

type RegistrationData struct {
	FullName      string
	Group         string
	PendingStatus string
}

type User struct {
	TelegramID int64
	Username   string
	FullName   string
	Group      string
}

type Attendance struct {
	Date       string
	Time       string
	TelegramID int64
	Username   string
	FullName   string
	Group      string
	Status     string
	Reason     string
}
