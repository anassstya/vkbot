package handler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	botgolang "github.com/mail-ru-im/bot-golang"
)

type EventDraft struct {
	State       string
	Title       string
	Description string
	Department  string
	Gender      string
	TimeToSend  time.Time
}

func (h *Handler) createEvent(ctx context.Context, chatID string) {
	role, err := h.UserRepo.GetRole(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения роли %s: %v", chatID, err)
		return
	}

	if role == "admin" {
		h.Drafts[chatID] = &EventDraft{State: "waiting_title"}

		msg := h.Bot.NewTextMessage(chatID,
			"Введите название события:")
		msg.Send()

	} else {
		h.Bot.NewTextMessage(chatID, "Команда доступна только администратору").Send()
		return
	}

}

func (h *Handler) handleEventDraft(ctx context.Context, chatID, text string, draft *EventDraft) {
	switch draft.State {
	case "waiting_title":
		draft.Title = text
		draft.State = "waiting_description"
		h.Bot.NewTextMessage(chatID, "Введите описание события:").Send()

	case "waiting_description":
		draft.Description = text
		draft.State = "waiting_department"

		h.SendWithButtons(chatID, "Введите отдел:",
			[][]botgolang.Button{
				{
					{Text: "💻 IT", CallbackData: "deptEvent:it"},
					{Text: "📊 Аналитика", CallbackData: "deptEvent:analytics"},
				},
				{
					{Text: "🎨 Дизайн", CallbackData: "deptEvent:design"},
					{Text: "📖 Менеджмент", CallbackData: "deptEvent:management"},
				},
				{
					{Text: "Все отделы", CallbackData: "deptEvent:all"},
				},
			})
		return

	case "waiting_time_to_send":
		t, err := time.Parse("2006-01-02 15:04", text)
		if err != nil {
			h.Bot.NewTextMessage(chatID, "Неверный формат даты. Попробуйте ещё раз: 2006-01-02 15:04").Send()
			return
		}

		now := time.Now()
		if !t.After(now) {
			h.Bot.NewTextMessage(chatID, "Дата отправки должна быть позже текущего времени").Send()
			return
		}

		draft.TimeToSend = t
		draft.State = "ready_to_send"

		h.Bot.NewTextMessage(chatID,
			"Название: "+draft.Title+"\n"+
				"Описание: "+draft.Description+"\n"+
				"Отдел: "+draft.Department+"\n"+
				"Пол: "+draft.Gender+"\n"+
				"Дата: "+draft.TimeToSend.Format("2006-01-02 15:04")).Send()

		h.SendWithButtons(chatID, "Подтвердить?", [][]botgolang.Button{
			{{Text: "✅ Создать", CallbackData: "readyEvent:yes"}},
			{{Text: "❌ Отмена", CallbackData: "readyEvent:no"}},
		})

	}
}

func (h *Handler) SendEvent(ctx context.Context, draft *EventDraft, chatID string) error {
	err := h.UserRepo.AddEvent(
		ctx,
		draft.Title,
		draft.Description,
		draft.Department,
		draft.Gender,
		chatID,
		draft.TimeToSend,
	)

	return err
}

func (h *Handler) GetAllEvents(ctx context.Context, chatID string) {
	events, err := h.UserRepo.GetEvents(ctx, chatID)

	if err != nil {
		msg := h.Bot.NewTextMessage(chatID, "Не удалось получить список событий. Попробуйте позже.")
		if err := msg.Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}
		return
	}

	if len(events) == 0 {
		msg := h.Bot.NewTextMessage(chatID, "📭 Событий пока нет.")
		if err := msg.Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}
		return
	}

	var builder strings.Builder
	builder.WriteString("📅 Ваши события:\n\n")

	for i, event := range events {
		builder.WriteString(fmt.Sprintf(
			"%d. %s\nОписание: %s\nОтдел: %s\nПол: %s\nДата: %s\nСтатус: %s\n\n",
			i+1,
			event.Title,
			event.Description,
			event.Department,
			event.Gender,
			event.TimeToSend.Format("2006-01-02 15:04"),
			event.Status,
		))
	}

	msg := h.Bot.NewTextMessage(chatID, builder.String())
	if err := msg.Send(); err != nil {
		log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
	}
}
