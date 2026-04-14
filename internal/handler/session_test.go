package handler

import (
	"context"
	"testing"
	"time"
)

func TestSession_NewSession_InitialState(t *testing.T) {
	chatID := "test_user_123"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)

	if session == nil {
		t.Fatal("NewSession вернул nil")
	}
	if session.ChatID != chatID {
		t.Errorf("ожидали ChatID=%s, получили %s", chatID, session.ChatID)
	}
	if session.State != stateIdle {
		t.Errorf("ожидали начальное состояние stateIdle, получили %s", session.State)
	}
}

func TestSession_Handle_WaitingTitle(t *testing.T) {
	ctx := context.Background()
	chatID := "title_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingTitle

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "Важное уведомление о собрании",
		Name:   "Анна",
	}

	session.Handle(ctx, event)

	if session.Draft.Title != "Важное уведомление о собрании" {
		t.Errorf("ожидали Draft.Title='Важное уведомление о собрании', получили '%s'", session.Draft.Title)
	}

	if session.State != stateWaitingDescription {
		t.Errorf("ожидали состояние stateWaitingDescription, получили %s", session.State)
	}
}

func TestSession_HandleCallback_RoleEmployee(t *testing.T) {
	ctx := context.Background()
	chatID := "callback_role_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)

	event := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "role:employee",
		Name:      "Тест",
		MessageID: "msg_123",
	}

	session.Handle(ctx, event)

	calls := mockRepo.UpdateRoleCalls()

	if len(calls) != 1 {
		t.Errorf("ожидали 1 вызов UpdateRole, получили: %d", len(calls))
	} else if calls[0].role != "employee" {
		t.Errorf("ожидали роль 'employee', получили: %s", calls[0].role)
	} else if calls[0].chatID != chatID {
		t.Errorf("ожидали chatID '%s', получили: %s", chatID, calls[0].chatID)
	}
}

func TestSession_Handle_WaitingTimeToSend_InvalidFormat(t *testing.T) {
	ctx := context.Background()
	chatID := "date_validation_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingTimeToSend

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "это-не-дата",
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	if session.State != stateWaitingTimeToSend {
		t.Errorf("ожидали, что состояние останется stateWaitingTimeToSend, получили %s", session.State)
	}

	if !session.Draft.TimeToSend.IsZero() {
		t.Errorf("ожидали, что TimeToSend не будет установлен, но получили %v", session.Draft.TimeToSend)
	}
}

func TestSession_Handle_WaitingTimeToSend_PastDate(t *testing.T) {
	ctx := context.Background()
	chatID := "past_date_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingTimeToSend

	pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04")

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   pastDate,
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	if session.State != stateWaitingTimeToSend {
		t.Errorf("ожидали, что состояние останется stateWaitingTimeToSend, получили %s", session.State)
	}

	if !session.Draft.TimeToSend.IsZero() {
		t.Errorf("ожидали, что TimeToSend не будет установлен для даты в прошлом")
	}
}

func TestSession_Handle_Cancel(t *testing.T) {
	ctx := context.Background()
	chatID := "cancel_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingTitle // ← сессия "занята"

	session.Draft.Title = "Какой-то старый заголовок"
	session.Draft.Description = "Старое описание"

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "/cancel",
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	if session.State != stateIdle {
		t.Errorf("ожидали состояние stateIdle после отмены, получили %s", session.State)
	}

	if session.Draft.Title != "" {
		t.Errorf("ожидали пустой заголовок после отмены, получили '%s'", session.Draft.Title)
	}
}

func TestSession_HandleCallback_DepartmentSelection(t *testing.T) {
	ctx := context.Background()
	chatID := "dept_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingDepartment
	event := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "deptNotification:it",
		Name:      "Тест",
		MessageID: "msg_456",
	}

	session.Handle(ctx, event)

	if session.Draft.RecipientsDepartment != "it" {
		t.Errorf("ожидали отдел 'it', получили '%s'", session.Draft.RecipientsDepartment)
	}

	if session.State != stateWaitingGender {
		t.Errorf("ожидали состояние stateWaitingGender, получили %s", session.State)
	}
}

func TestSession_HandleCallback_ConfirmSave(t *testing.T) {
	ctx := context.Background()
	chatID := "confirm_save_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateReadyToSend

	session.Draft.Title = "Корпоратив"
	session.Draft.Description = "Сбор в 18:00"
	session.Draft.RecipientsDepartment = "all"
	session.Draft.RecipientsGender = "all"
	session.Draft.Type = "one_time"
	session.Draft.TimeToSend = time.Now().Add(24 * time.Hour)

	event := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "readyNotification:yes",
		Name:      "Тест",
		MessageID: "msg_789",
	}

	session.Handle(ctx, event)

	if count := mockRepo.AddNotificationCallCount(); count != 1 {
		t.Errorf("ожидали 1 вызов AddNotification, получили: %d", count)
	}

	if session.State != stateIdle {
		t.Errorf("ожидали состояние stateIdle после сохранения, получили %s", session.State)
	}

	if session.Draft.Title != "" {
		t.Errorf("ожидали пустой заголовок после сохранения, получили '%s'", session.Draft.Title)
	}
}

func TestSession_HandleCallback_ConfirmReject(t *testing.T) {
	ctx := context.Background()
	chatID := "reject_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateReadyToSend
	session.Draft.Title = "Тестовое уведомление"

	event := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "readyNotification:no",
		Name:      "Тест",
		MessageID: "msg_final",
	}

	session.Handle(ctx, event)

	if count := mockRepo.AddNotificationCallCount(); count != 0 {
		t.Errorf("ожидали 0 вызовов AddNotification при отмене, получили: %d", count)
	}

	if session.State != stateIdle {
		t.Errorf("ожидали stateIdle, получили %s", session.State)
	}

	if session.Draft.Title != "" {
		t.Error("черновик должен быть очищен при отказе")
	}
}

func TestSession_HandleMessage_Event_ChecksRole(t *testing.T) {
	ctx := context.Background()
	chatID := "event_role_test"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateIdle

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "/event",
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	roleCalls := mockRepo.GetRoleCalls()

	if len(roleCalls) != 1 {
		t.Errorf("ожидали 1 вызов GetRole, получили: %d", len(roleCalls))
	} else if roleCalls[0] != chatID {
		t.Errorf("ожидали проверку роли для chatID=%s, получили %s", chatID, roleCalls[0])
	}
}

func TestSession_HandleMessage_Event_NotAdmin(t *testing.T) {
	ctx := context.Background()
	chatID := "event_employee_test"
	mockBot := &MockBot{}

	mockRepo := &MockUserRepo{
		getRoleFunc: func(ctx context.Context, chatID string) (string, error) {
			return "employee", nil // ← не админ!
		},
	}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateIdle

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "/event",
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	if session.State != stateIdle {
		t.Errorf("ожидали, что состояние останется stateIdle при отказе в доступе, получили %s", session.State)
	}

	if len(mockRepo.GetRoleCalls()) != 1 {
		t.Error("ожидали вызов GetRole для проверки прав")
	}
}

func TestSession_HandleCallback_GenderSelection(t *testing.T) {
	ctx := context.Background()
	chatID := "gender_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingGender

	event := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "genderNotification:female",
		Name:      "Тест",
		MessageID: "msg_gender_123",
	}

	session.Handle(ctx, event)

	if session.Draft.RecipientsGender != "female" {
		t.Errorf("ожидали гендер 'female', получили '%s'", session.Draft.RecipientsGender)
	}

	if session.State != stateWaitingNotificationType {
		t.Errorf("ожидали состояние stateWaitingNotificationType, получили %s", session.State)
	}
}

func TestSession_HandleMessage_Profile(t *testing.T) {
	ctx := context.Background()
	chatID := "profile_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateIdle

	event := Event{
		ChatID: chatID,
		Type:   "message",
		Text:   "/profile",
		Name:   "Тест",
	}

	session.Handle(ctx, event)

	calls := mockRepo.GetUserCalls()

	if len(calls) != 1 {
		t.Errorf("ожидали 1 вызов GetUser, получили: %d", len(calls))
	} else if calls[0] != chatID {
		t.Errorf("ожидали запрос для chatID=%s, получили %s", chatID, calls[0])
	}

	if session.State != stateIdle {
		t.Errorf("ожидали, что состояние останется stateIdle, получили %s", session.State)
	}
}

func TestSession_HandleCallback_RecurringNotification(t *testing.T) {
	ctx := context.Background()
	chatID := "recurring_test_user"
	mockBot := &MockBot{}
	mockRepo := &MockUserRepo{}

	session := NewSession(chatID, mockBot, mockRepo)
	session.State = stateWaitingNotificationType

	session.Draft.Title = "Еженедельный отчёт"
	session.Draft.Description = "Напоминание о задачах"
	session.Draft.RecipientsDepartment = "it"
	session.Draft.RecipientsGender = "all"

	event1 := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "notificationType:regular",
		Name:      "Тест",
		MessageID: "msg_type_123",
	}
	session.Handle(ctx, event1)

	if session.Draft.Type != "recurring" {
		t.Errorf("ожидали тип 'recurring', получили '%s'", session.Draft.Type)
	}

	if session.State != stateWaitingRecurrence {
		t.Errorf("ожидали состояние stateWaitingRecurrence, получили %s", session.State)
	}

	event2 := Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      "recurrence:weekly",
		Name:      "Тест",
		MessageID: "msg_rec_456",
	}
	session.Handle(ctx, event2)

	if session.Draft.Recurrence != "weekly" {
		t.Errorf("ожидали повторение 'weekly', получили '%s'", session.Draft.Recurrence)
	}
	if session.State != stateWaitingRecTime {
		t.Errorf("ожидали состояние stateWaitingRecTime, получили %s", session.State)
	}
}
