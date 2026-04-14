package handler

import (
	"log"

	botgolang "github.com/mail-ru-im/bot-golang"
)

func (h *Handler) SendWithButtons(chatID, text string, buttons [][]botgolang.Button) {
	keyboard := botgolang.NewKeyboard()

	for _, row := range buttons {
		keyboard.AddRow(row...)
	}

	msg := h.Bot.NewInlineKeyboardMessage(chatID, text, keyboard)

	if err := msg.Send(); err != nil {
		log.Printf("Ошибка отправки с кнопками %s: %v", chatID, err)
	}
}

func (h *Handler) replaceButtons(chatID, messageID, text string) {
	edited := h.Bot.NewTextMessage(chatID, text)
	edited.ID = messageID

	if err := edited.Edit(); err != nil {
		log.Printf("Ошибка редактирования сообщения: %v", err)
	}
}

func translateDeptAndGender(dept, gender string) (string, string) {
	switch gender {
	case "all":
		gender = "для всех"
	case "male":
		gender = "мужской"
	case "female":
		gender = "женский"
	}

	switch dept {
	case "all":
		dept = "все отделы"
	case "it":
		dept = "IT отдел"
	case "design":
		dept = "дизайна"
	case "management":
		dept = "менеджмент и реклама"
	case "analytics":
		dept = "аналитика"
	}

	return dept, gender
}
