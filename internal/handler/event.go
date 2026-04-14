package handler

type Event struct {
	ChatID    string
	Type      string // "message" | "callback"
	Text      string
	Data      string
	Name      string
	MessageID string
}
