package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anassstya/vkbot/internal/repository"
	botgolang "github.com/mail-ru-im/bot-golang"
)

type Scheduler struct {
	UserRepo UserRepoInterface // ← интерфейс
	Bot      BotInterface      // ← интерфейс
}

func NewScheduler(repo UserRepoInterface, bot BotInterface) *Scheduler {
	return &Scheduler{
		UserRepo: repo,
		Bot:      bot,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	log.Println("Scheduler запущен")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler остановлен")
			return
		case <-ticker.C:
			s.processReadyNotifications(ctx)
		}
	}
}

func (s *Scheduler) processReadyNotifications(ctx context.Context) {
	notifications, err := s.UserRepo.GetReadyNotifications(ctx)
	if err != nil {
		log.Printf("❌ Ошибка получения ready notifications: %v", err)
		return
	}

	if len(notifications) == 0 {
		return
	}

	log.Printf("📨 Найдено %d уведомлений", len(notifications))

	for _, notification := range notifications {
		if err := s.DispatchNotification(ctx, notification); err != nil {
			log.Printf("💥 Ошибка уведомления %d: %v", notification.ID, err)
		}
	}
}
func (s *Scheduler) DispatchNotification(ctx context.Context, notification repository.Notification) error {
	recipients, err := s.UserRepo.GetRecipientsWithInfo(
		ctx,
		notification.RecipientsDepartment,
		notification.RecipientsGender,
	)
	if err != nil {
		return fmt.Errorf("GetRecipientsWithInfo: %w", err)
	}

	log.Printf("📨 Уведомление %d: %d получателей", notification.ID, len(recipients))

	if len(recipients) == 0 {
		return s.UserRepo.UpdateNotificationStatus(ctx, notification.ID, "no_recipients")
	}

	sent := 0
	failed := 0

	for _, rec := range recipients {

		title := notification.Title
		desc := strings.ReplaceAll(notification.Description, "{name}", rec.Name)

		text := fmt.Sprintf(
			"📢 %s\n\n%s, %s!\n%s",
			title,
			genderGreeting(rec.Gender),
			rec.Name,
			desc,
		)

		keyboard := botgolang.NewKeyboard()

		keyboard.AddRow(
			botgolang.Button{
				Text:         "✅ Прочитал",
				CallbackData: fmt.Sprintf("read:%d", notification.ID),
			},
		)

		msg := s.Bot.NewInlineKeyboardMessage(rec.ChatID, text, keyboard)

		if err := msg.Send(); err != nil {
			log.Printf("❌ Ошибка отправки в %s: %v", rec.ChatID, err)
			failed++
		} else {
			sent++
		}
	}

	log.Printf("✅ %d/%d отправлено", sent, len(recipients))

	if notification.Type == "one_time" {
		status := "sent"
		if sent == 0 {
			status = "failed"
		} else if failed > 0 {
			status = "partial"
		}
		return s.UserRepo.UpdateNotificationStatus(ctx, notification.ID, status)
	}

	if notification.Type == "recurring" {
		if sent == 0 {
			return s.UserRepo.UpdateNotificationStatus(ctx, notification.ID, "failed")
		}
		return s.UserRepo.MarkRecurringNotificationSent(ctx, notification)
	}

	if notification.Type == "trigger" {
		if sent == 0 {
			return s.UserRepo.UpdateNotificationStatus(ctx, notification.ID, "failed")
		}
		return s.UserRepo.MarkTriggerNotificationSent(ctx, notification.ID)
	}

	return nil
}

func genderGreeting(gender string) string {
	switch gender {
	case "male":
		return "Уважаемый"
	case "female":
		return "Уважаемая"
	default:
		return "Уважаемый(-ая)"
	}
}
