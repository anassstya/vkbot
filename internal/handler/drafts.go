package handler

import (
	"time"
)

type NotificationDraft struct {
	State                string
	Title                string
	Description          string
	RecipientsDepartment string
	RecipientsGender     string
	Type                 string
	Recurrence           string
	TimeToSend           time.Time
	NextSend             time.Time
}
