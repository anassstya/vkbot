// scheduler/interface.go
package scheduler

import (
	"context"

	"github.com/anassstya/vkbot/internal/repository"
	botgolang "github.com/mail-ru-im/bot-golang"
)

// BotInterface — интерфейс для бота в планировщике
// Содержит только те методы, которые реально используются
type BotInterface interface {
	NewTextMessage(chatID, text string) *botgolang.Message
	NewInlineKeyboardMessage(chatID, text string, keyboard botgolang.Keyboard) *botgolang.Message
}

// UserRepoInterface — интерфейс репозитория для планировщика
// Можно скопировать из handler или импортировать, если он экспортирован
type UserRepoInterface interface {
	GetReadyNotifications(ctx context.Context) ([]repository.Notification, error)
	GetRecipientsWithInfo(ctx context.Context, department, gender string) ([]repository.Recipient, error)
	UpdateNotificationStatus(ctx context.Context, notificationID int, status string) error
	MarkRecurringNotificationSent(ctx context.Context, notification repository.Notification) error
	MarkTriggerNotificationSent(ctx context.Context, notificationID int) error
}
