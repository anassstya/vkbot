// scheduler/mocks_test.go
package scheduler

import (
	"context"
	"sync"

	"github.com/anassstya/vkbot/internal/repository"
	botgolang "github.com/mail-ru-im/bot-golang"
)

// ===== Mock для UserRepoInterface =====
type MockUserRepo struct {
	GetReadyNotificationsFunc func(ctx context.Context) ([]repository.Notification, error)
	GetRecipientsFunc         func(ctx context.Context, dept, gender string) ([]repository.Recipient, error)
	UpdateStatusFunc          func(ctx context.Context, id int, status string) error
	MarkRecurringFunc         func(ctx context.Context, n repository.Notification) error

	mu sync.Mutex
	// Для отслеживания вызовов
	readyNotificationsCalls int
	updateStatusCalls       []struct {
		id     int
		status string
	}
}

func (m *MockUserRepo) GetReadyNotifications(ctx context.Context) ([]repository.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readyNotificationsCalls++

	if m.GetReadyNotificationsFunc != nil {
		return m.GetReadyNotificationsFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserRepo) GetRecipientsWithInfo(ctx context.Context, department, gender string) ([]repository.Recipient, error) {
	if m.GetRecipientsFunc != nil {
		return m.GetRecipientsFunc(ctx, department, gender)
	}
	// Возвращаем тестового получателя по умолчанию
	return []repository.Recipient{
		{ChatID: "test_user_123", Name: "Тест", Gender: "female"},
	}, nil
}

func (m *MockUserRepo) UpdateNotificationStatus(ctx context.Context, notificationID int, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateStatusCalls = append(m.updateStatusCalls, struct {
		id     int
		status string
	}{id: notificationID, status: status})

	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, notificationID, status)
	}
	return nil
}

func (m *MockUserRepo) MarkRecurringNotificationSent(ctx context.Context, notification repository.Notification) error {
	if m.MarkRecurringFunc != nil {
		return m.MarkRecurringFunc(ctx, notification)
	}
	return nil
}

func (m *MockUserRepo) MarkTriggerNotificationSent(ctx context.Context, notificationID int) error {
	return nil
}

// Вспомогательные методы для проверок в тестах
func (m *MockUserRepo) ReadyNotificationsCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readyNotificationsCalls
}

func (m *MockUserRepo) UpdateStatusCalls() []struct {
	id     int
	status string
} {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]struct {
		id     int
		status string
	}, len(m.updateStatusCalls))
	copy(calls, m.updateStatusCalls)
	return calls
}

// ===== Mock для BotInterface =====
type MockBot struct {
	sentMessages []struct {
		chatID string
		text   string
	}
	mu sync.Mutex
}

func (m *MockBot) NewTextMessage(chatID, text string) *botgolang.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, struct {
		chatID string
		text   string
	}{chatID: chatID, text: text})
	return &botgolang.Message{}
}

func (m *MockBot) NewInlineKeyboardMessage(chatID, text string, keyboard botgolang.Keyboard) *botgolang.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, struct {
		chatID string
		text   string
	}{chatID: chatID, text: text})
	return &botgolang.Message{}
}

// Вспомогательный метод для проверок
func (m *MockBot) SentMessagesCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sentMessages)
}
