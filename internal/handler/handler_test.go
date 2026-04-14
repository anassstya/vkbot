package handler

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func waitShort() {
	time.Sleep(50 * time.Millisecond)
}

func TestNewHandler(t *testing.T) {
	mockRepo := &MockUserRepo{}

	h := NewHandler(mockRepo, nil)

	if h == nil {
		t.Fatal("NewHandler вернул nil")
	}
	if h.UserRepo == nil {
		t.Error("UserRepo не инициализирован")
	}

	if h.sessions == nil {
		t.Error("sessions map не инициализирован")
	}
}

func TestHandler_Handle_CallsAddUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepo{}

	h := NewHandler(mockRepo, nil)

	chatID := "12345"
	name := "Анна"
	text := "Привет!"

	h.Handle(ctx, chatID, name, text)

	waitShort()

	if count := mockRepo.AddUserCallCount(); count != 1 {
		t.Errorf("ожидали 1 вызов AddUser, получили: %d", count)
	}
}

func TestHandler_dispatch_NoPanic(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepo{}

	h := NewHandler(mockRepo, nil)

	chatID := "test_chat_999"
	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "test",
		Name:   "TestUser",
	}

	h.dispatch(ctx, event)

	waitShort()
}

func TestHandler_Handle_AddUserError(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepo{
		AddUserFunc: func(ctx context.Context, chatID, name string) error {
			return fmt.Errorf("база недоступна")
		},
	}

	h := NewHandler(mockRepo, nil)

	h.Handle(ctx, "123", "Тест", "привет")

	if count := mockRepo.AddUserCallCount(); count != 1 {
		t.Errorf("ожидали 1 вызов AddUser, получили: %d", count)
	}
}

func TestHandler_HandleCallback_Basic(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepo{}

	h := NewHandler(mockRepo, nil)

	chatID := "callback_user_123"
	name := "КнопкаТест"
	data := "deptNotification:it"
	messageID := "msg_999"

	h.HandleCallback(ctx, chatID, name, data, messageID)

	waitShort()

	h.mu.Lock()
	_, exists := h.sessions[chatID]
	h.mu.Unlock()

	if !exists {
		t.Errorf("ожидали, что сессия для chatID=%s будет создана", chatID)
	}
	if count := mockRepo.AddUserCallCount(); count != 0 {
		t.Errorf("ожидали 0 вызовов AddUser для колбэка, получили: %d", count)
	}
}

func TestHandler_Handle_Commands(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		wantAddUser int
	}{
		{"команда /start", "/start", 1},
		{"команда /info", "/info", 1},
		{"команда /event", "/event", 1},
		{"простое сообщение", "Привет, бот!", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := &MockUserRepo{}
			mockBot := &MockBot{}

			h := NewHandler(mockRepo, mockBot)

			h.Handle(ctx, "test_chat_"+tt.name, "Tester", tt.command)
			waitShort()

			if got := mockRepo.AddUserCallCount(); got != tt.wantAddUser {
				t.Errorf("AddUser() вызван %d раз, ожидалось %d", got, tt.wantAddUser)
			}
		})
	}
}

func TestSession_HandleMessage_MyEvents(t *testing.T) {
	ctx := context.Background()
	chatID := "my_events_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateIdle

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "/my_events",
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	calls := mockRepo.GetMyNotificationsCalls()

	if len(calls) != 1 {
		t.Errorf("ожидали 1 вызов GetMyNotifications, получили: %d", len(calls))
	} else if calls[0] != chatID {
		t.Errorf("ожидали запрос для chatID=%s, получили %s", chatID, calls[0])
	}

	if session.State != stateIdle {
		t.Errorf("ожидали, что состояние останется stateIdle, получили %s", session.State)
	}
}

func TestSession_GetNotificationByID(t *testing.T) {
	ctx := context.Background()
	chatID := "get_by_id_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	NewSession(chatID, mockBot, mockRepo)

	notificationID := 42

	notification, err := mockRepo.GetNotificationByID(ctx, notificationID)

	if err != nil {
		t.Errorf("ожидали отсутствие ошибки, получили: %v", err)
	}
	if notification.ID != notificationID {
		t.Errorf("ожидали ID=%d, получили %d", notificationID, notification.ID)
	}

	calls := mockRepo.GetNotificationByIDCalls()
	if len(calls) != 1 || calls[0] != notificationID {
		t.Errorf("ожидали вызов для ID=%d, получили: %v", notificationID, calls)
	}
}

func TestSession_HandleCallback_MarkAsRead(t *testing.T) {
	ctx := context.Background()
	chatID := "mark_read_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateIdle

	event := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "read:42",
		Name:      "Тест",
		MessageID: "msg_read_123",
	}

	session.Handle(ctx, event)

	calls := mockRepo.MarkAsReadCalls()

	if len(calls) != 1 {
		t.Errorf("ожидали 1 вызов MarkAsRead, получили: %d", len(calls))
	} else {
		if calls[0].notificationID != 42 {
			t.Errorf("ожидали notificationID=42, получили %d", calls[0].notificationID)
		}
		if calls[0].chatID != chatID {
			t.Errorf("ожидали chatID=%s, получили %s", chatID, calls[0].chatID)
		}
	}
}
