package handler

import (
	"context"
	"log"
	"sync"
)

type Handler struct {
	UserRepo UserRepoInterface
	Bot      BotInterface

	sessions map[string]chan Event
	mu       sync.Mutex
}

type TextMessageInterface interface {
	Send() error
	Edit() error
}

func NewHandler(userRepo UserRepoInterface, bot BotInterface) *Handler {
	return &Handler{
		UserRepo: userRepo,
		Bot:      bot,
		sessions: make(map[string]chan Event),
	}
}

func (h *Handler) Handle(ctx context.Context, chatID, name, text string) {
	if err := h.UserRepo.AddUser(ctx, chatID, name); err != nil {
		log.Printf("Ошибка сохранения пользователя %s: %v", chatID, err)
	}

	h.dispatch(ctx, Event{
		ChatID: chatID,
		Type:   "message",
		Text:   text,
		Name:   name,
	})
}

func (h *Handler) HandleCallback(ctx context.Context, chatID, name, data, messageID string) {
	h.dispatch(ctx, Event{
		ChatID:    chatID,
		Type:      "callback",
		Data:      data,
		Name:      name,
		MessageID: messageID,
	})
}

func (h *Handler) dispatch(ctx context.Context, event Event) {
	h.mu.Lock()
	ch, ok := h.sessions[event.ChatID]
	if !ok {
		ch = make(chan Event, 100)
		h.sessions[event.ChatID] = ch
		go h.runSession(ctx, event.ChatID, ch)
	}
	h.mu.Unlock()

	ch <- event
}

func (h *Handler) runSession(ctx context.Context, chatID string, ch <-chan Event) {
	session := NewSession(chatID, h.Bot, h.UserRepo)

	for event := range ch {
		session.Handle(ctx, event)
	}
}
