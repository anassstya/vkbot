// scheduler/scheduler_test.go
package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/anassstya/vkbot/internal/repository"
)

// ===== Тест 1: Если отправка не удалась — статус меняется на "failed" =====
func TestScheduler_DispatchNotification_SendFailed(t *testing.T) {
	// Подготовка
	mockRepo := &MockUserRepo{}
	mockBot := &MockBot{}

	s := NewScheduler(mockRepo, mockBot)

	// Создаём тестовое уведомление, которое «пора отправлять»
	notification := repository.Notification{
		ID:                   42,
		Title:                "Тестовое уведомление",
		Description:          "Проверка работы планировщика",
		RecipientsDepartment: "it",
		RecipientsGender:     "all",
		Type:                 "one_time",
		TimeToSend:           ptrTime(time.Now().Add(-1 * time.Minute)),
	}

	// Действие
	// Вызываем метод напрямую
	_ = s.DispatchNotification(context.Background(), notification)

	// Проверка:
	// Поскольку мок возвращает &botgolang.Message{} без клиента,
	// реальная библиотека вернёт ошибку при Send().
	// Проверяем, что Scheduler корректно обработал это и поставил статус "failed".

	statusCalls := mockRepo.UpdateStatusCalls()

	if len(statusCalls) != 1 {
		t.Errorf("ожидали 1 вызов UpdateNotificationStatus, получили: %d", len(statusCalls))
	} else if statusCalls[0].id != 42 {
		t.Errorf("ожидали обновление статуса для notification ID=42, получили %d", statusCalls[0].id)
	} else if statusCalls[0].status != "failed" { // ← ожидаем "failed", а не "sent"
		t.Errorf("ожидали статус 'failed' при ошибке отправки, получили '%s'", statusCalls[0].status)
	}
}

// Вспомогательная функция для указателей на time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}

// ===== Тест 2: Если нет получателей — статус "no_recipients" =====
func TestScheduler_DispatchNotification_NoRecipients(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetRecipientsFunc: func(ctx context.Context, dept, gender string) ([]repository.Recipient, error) {
			return nil, nil // ← нет получателей
		},
	}
	mockBot := &MockBot{}
	s := NewScheduler(mockRepo, mockBot)

	notification := repository.Notification{
		ID: 43, Type: "one_time", TimeToSend: ptrTime(time.Now().Add(-1 * time.Minute)),
		RecipientsDepartment: "it", RecipientsGender: "all",
	}

	_ = s.DispatchNotification(context.Background(), notification)

	statusCalls := mockRepo.UpdateStatusCalls()
	if len(statusCalls) != 1 || statusCalls[0].status != "no_recipients" {
		t.Errorf("ожидали статус 'no_recipients', получили: %+v", statusCalls)
	}
}
