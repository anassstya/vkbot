package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anassstya/vkbot/internal/repository"
	botgolang "github.com/mail-ru-im/bot-golang"
)

type MockUserRepo struct {
	AddUserFunc func(ctx context.Context, chatID, name string) error
	calls       []struct {
		chatID string
		name   string
	}

	mu                   sync.Mutex
	addNotificationCalls int

	updateRoleCalls []struct {
		role   string
		chatID string
	}
	getRoleCalls []struct {
		chatID string
	}
	getRoleFunc func(ctx context.Context, chatID string) (string, error)

	addDeptCalls []struct {
		dept   string
		chatID string
	}

	addGenderCalls []struct {
		gender string
		chatID string
	}

	getUserCalls []string
	getUserFunc  func(ctx context.Context, chatID string) (repository.User, error)

	getMyNotificationsCalls  []string
	getNotificationByIDCalls []int

	getNotificationStatsCalls  int
	getReadyNotificationsCalls int

	markAsReadCalls []struct {
		notificationID int
		chatID         string
	}

	markRecurringCalls []int
}

func (m *MockUserRepo) AddUser(ctx context.Context, chatID, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, struct{ chatID, name string }{chatID: chatID, name: name})
	if m.AddUserFunc != nil {
		return m.AddUserFunc(ctx, chatID, name)
	}
	return nil
}

func (m *MockUserRepo) AddUserCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *MockUserRepo) GetUser(ctx context.Context, chatID string) (repository.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getUserCalls = append(m.getUserCalls, chatID)

	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, chatID)
	}

	return repository.User{
		ChatID:     chatID,
		Name:       "Тестовый Пользователь",
		Role:       "employee",
		Gender:     "female",
		Department: "it",
		CreatedAt:  time.Now().AddDate(-1, 0, 0), // год назад
	}, nil
}

func (m *MockUserRepo) GetUserCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]string, len(m.getUserCalls))
	copy(calls, m.getUserCalls)
	return calls
}

func (m *MockUserRepo) UpdateRole(ctx context.Context, role, chatID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updateRoleCalls = append(m.updateRoleCalls, struct {
		role   string
		chatID string
	}{role: role, chatID: chatID})

	return nil
}

func (m *MockUserRepo) UpdateRoleCalls() []struct {
	role   string
	chatID string
} {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]struct {
		role   string
		chatID string
	}, len(m.updateRoleCalls))
	copy(calls, m.updateRoleCalls)
	return calls
}

func (m *MockUserRepo) AddDept(ctx context.Context, dept, chatID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addDeptCalls = append(m.addDeptCalls, struct {
		dept   string
		chatID string
	}{dept: dept, chatID: chatID})

	return nil
}

func (m *MockUserRepo) AddDeptCalls() []struct {
	dept   string
	chatID string
} {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]struct {
		dept   string
		chatID string
	}, len(m.addDeptCalls))
	copy(calls, m.addDeptCalls)
	return calls
}

func (m *MockUserRepo) AddGender(ctx context.Context, gender, chatID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addGenderCalls = append(m.addGenderCalls, struct {
		gender string
		chatID string
	}{gender: gender, chatID: chatID})

	return nil
}

func (m *MockUserRepo) AddGenderCalls() []struct {
	gender string
	chatID string
} {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]struct {
		gender string
		chatID string
	}, len(m.addGenderCalls))
	copy(calls, m.addGenderCalls)
	return calls
}

func (m *MockUserRepo) GetRole(ctx context.Context, chatID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getRoleCalls = append(m.getRoleCalls, struct{ chatID string }{chatID: chatID})

	if m.getRoleFunc != nil {
		return m.getRoleFunc(ctx, chatID)
	}
	return "admin", nil
}

func (m *MockUserRepo) GetRoleCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]string, len(m.getRoleCalls))
	for i, c := range m.getRoleCalls {
		calls[i] = c.chatID
	}
	return calls
}

func (m *MockUserRepo) AddNotification(ctx context.Context, title, description, recipientsDepartment, recipientsGender, createdBy string, timeToSend time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addNotificationCalls++
	return 1, nil
}

func (m *MockUserRepo) AddNotificationCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.addNotificationCalls
}

func (m *MockUserRepo) AddRecurringNotification(ctx context.Context, title, description, recipientsDepartment, recipientsGender, createdBy, recurrence string, nextSend time.Time) (int, error) {
	return 0, nil
}

func (m *MockUserRepo) GetMyNotifications(ctx context.Context, createdBy string) ([]repository.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getMyNotificationsCalls = append(m.getMyNotificationsCalls, createdBy)

	return []repository.Notification{
		{
			ID:          1,
			Title:       "Тестовое уведомление 1",
			Description: "Описание теста",
			Status:      "sent",
			Type:        "one_time",
			CreatedAt:   time.Now().AddDate(0, 0, -1),
		},
		{
			ID:          2,
			Title:       "Тестовое уведомление 2",
			Description: "Ещё один тест",
			Status:      "scheduled",
			Type:        "recurring",
			CreatedAt:   time.Now(),
		},
	}, nil
}

func (m *MockUserRepo) GetMyNotificationsCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]string, len(m.getMyNotificationsCalls))
	copy(calls, m.getMyNotificationsCalls)
	return calls
}

func (m *MockUserRepo) GetNotificationByID(ctx context.Context, notificationID int) (repository.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getNotificationByIDCalls = append(m.getNotificationByIDCalls, notificationID)

	return repository.Notification{
		ID:          notificationID,
		Title:       fmt.Sprintf("Уведомление #%d", notificationID),
		Description: "Тестовое описание",
		Status:      "sent",
		Type:        "one_time",
		CreatedAt:   time.Now().AddDate(0, 0, -1),
	}, nil
}

func (m *MockUserRepo) GetNotificationByIDCalls() []int {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]int, len(m.getNotificationByIDCalls))
	copy(calls, m.getNotificationByIDCalls)
	return calls
}

func (m *MockUserRepo) GetReadyNotifications(ctx context.Context) ([]repository.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getReadyNotificationsCalls++

	return []repository.Notification{
		{
			ID:                   99,
			Title:                "Готовое к отправке",
			Description:          "Тестовое описание",
			RecipientsDepartment: "all",
			RecipientsGender:     "all",
			Type:                 "one_time",
			TimeToSend:           ptrTime(time.Now().Add(-1 * time.Minute)),
		},
	}, nil
}
func (m *MockUserRepo) GetNotificationStatsCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getNotificationStatsCalls
}

func (m *MockUserRepo) GetReadyNotificationsCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getReadyNotificationsCalls
}

func ptrTime(t time.Time) *time.Time { return &t }

func (m *MockUserRepo) GetNotificationStats(ctx context.Context) ([]repository.NotificationStats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getNotificationStatsCalls++

	return []repository.NotificationStats{
		{ID: 1, Title: "Тестовая рассылка", DeliveredCount: 10, ReadCount: 5, OpenRate: 50.0},
		{ID: 2, Title: "Вторая рассылка", DeliveredCount: 20, ReadCount: 15, OpenRate: 75.0},
	}, nil
}

func (m *MockUserRepo) UpdateNotificationStatus(ctx context.Context, notificationID int, status string) error {
	return nil
}

func (m *MockUserRepo) MarkAsRead(ctx context.Context, notificationID int, chatID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markAsReadCalls = append(m.markAsReadCalls, struct {
		notificationID int
		chatID         string
	}{notificationID: notificationID, chatID: chatID})
	return nil
}

func (m *MockUserRepo) MarkRecurringNotificationSent(ctx context.Context, notification repository.Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markRecurringCalls = append(m.markRecurringCalls, notification.ID)
	return nil
}

func (m *MockUserRepo) MarkAsReadCalls() []struct {
	notificationID int
	chatID         string
} {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]struct {
		notificationID int
		chatID         string
	}, len(m.markAsReadCalls))
	copy(calls, m.markAsReadCalls)
	return calls
}

func (m *MockUserRepo) MarkRecurringCalls() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]int, len(m.markRecurringCalls))
	copy(calls, m.markRecurringCalls)
	return calls
}

func (m *MockUserRepo) GetWelcomeTrigger(ctx context.Context) (repository.Notification, error) {
	return repository.Notification{}, nil
}

func (m *MockUserRepo) EnsureWelcomeTrigger(ctx context.Context) error { return nil }

func (m *MockUserRepo) MarkTriggerNotificationSent(ctx context.Context, notificationID int) error {
	return nil
}

func (m *MockUserRepo) GetRecipientsWithInfo(ctx context.Context, department, gender string) ([]repository.Recipient, error) {
	return nil, nil
}

type MockMessage struct{}

func (m *MockMessage) Send() error { return nil }
func (m *MockMessage) Edit() error { return nil }

type MockBot struct{}

func (m *MockBot) NewTextMessage(chatID, text string) *botgolang.Message {
	return &botgolang.Message{}
}

func (m *MockBot) NewInlineKeyboardMessage(chatID, text string, keyboard botgolang.Keyboard) *botgolang.Message {
	return &botgolang.Message{}
}
