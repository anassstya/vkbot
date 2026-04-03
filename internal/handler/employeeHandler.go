package handler

import (
	"context"
	"log"

	botgolang "github.com/mail-ru-im/bot-golang"
)

func (h *Handler) handleDept(ctx context.Context, chatID, messageID, dept, label string) {
	if err := h.UserRepo.AddDept(ctx, dept, chatID); err != nil {
		log.Printf("Ошибка обновления отдела %s: %v", chatID, err)
	}
	h.replaceButtons(chatID, messageID, "Выбери свой отдел:\n\nОтдел "+label+" выбран! ✅")

	h.SendWithButtons(chatID,
		"Теперь выберите ваш пол:",
		[][]botgolang.Button{
			{
				{Text: "Женский", CallbackData: "gender:female"},
				{Text: "Мужской", CallbackData: "gender:male"},
			},
		})
}

func (h *Handler) handleGender(ctx context.Context, chatID, messageID, gender, label string) {
	if err := h.UserRepo.AddGender(ctx, gender, chatID); err != nil {
		log.Printf("Ошибка обновления пола %s: %v", chatID, err)
	}

	h.replaceButtons(chatID, messageID, "Выберите ваш пол:\n\n"+label+" ✅")

	msg := h.Bot.NewTextMessage(chatID,
		"Профиль заполнен!\n\n"+
			"Теперь вы будете получать уведомления о корпоративных событиях вашего отдела.\n\n"+
			"Как только появится новое мероприятие — я сразу оповещу вас! 🔔\n\n"+
			"Если появятся вопросы — введите /info",
	)
	if err := msg.Send(); err != nil {
		log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
	}

}
