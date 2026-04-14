package handler

import (
	"context"
	"time"

	"github.com/anassstya/vkbot/internal/repository"
	botgolang "github.com/mail-ru-im/bot-golang"
)

type UserRepoInterface interface {
	AddUser(ctx context.Context, chatID, name string) error
	GetUser(ctx context.Context, chatID string) (repository.User, error)
	UpdateRole(ctx context.Context, role, chatID string) error
	AddDept(ctx context.Context, dept, chatID string) error
	AddGender(ctx context.Context, gender, chatID string) error
	GetRole(ctx context.Context, chatID string) (string, error)

	AddNotification(
		ctx context.Context,
		title, description, recipientsDepartment, recipientsGender, createdBy string,
		timeToSend time.Time,
	) (int, error)

	AddRecurringNotification(
		ctx context.Context,
		title, description, recipientsDepartment, recipientsGender, createdBy, recurrence string,
		nextSend time.Time,
	) (int, error)

	GetMyNotifications(ctx context.Context, createdBy string) ([]repository.Notification, error)
	GetNotificationByID(ctx context.Context, notificationID int) (repository.Notification, error)
	GetReadyNotifications(ctx context.Context) ([]repository.Notification, error)
	GetNotificationStats(ctx context.Context) ([]repository.NotificationStats, error)

	UpdateNotificationStatus(ctx context.Context, notificationID int, status string) error
	MarkAsRead(ctx context.Context, notificationID int, chatID string) error
	MarkRecurringNotificationSent(ctx context.Context, notification repository.Notification) error

	GetWelcomeTrigger(ctx context.Context) (repository.Notification, error)
	EnsureWelcomeTrigger(ctx context.Context) error
	MarkTriggerNotificationSent(ctx context.Context, notificationID int) error

	GetRecipientsWithInfo(ctx context.Context, department, gender string) ([]repository.Recipient, error)
}

type BotInterface interface {
	NewTextMessage(chatID, text string) *botgolang.Message
	NewInlineKeyboardMessage(chatID, text string, keyboard botgolang.Keyboard) *botgolang.Message
}
