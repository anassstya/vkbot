package handler

import (
	"context"
	"log"

	"github.com/anassstya/vkbot/internal/repository"
	botgolang "github.com/mail-ru-im/bot-golang"
)

type Handler struct {
	UserRepo *repository.UserRepo
	Bot      *botgolang.Bot
	Drafts   map[string]*EventDraft
}

func NewHandler(userRepo *repository.UserRepo, bot *botgolang.Bot) *Handler {
	return &Handler{
		UserRepo: userRepo,
		Bot:      bot,
		Drafts:   map[string]*EventDraft{},
	}
}

func (h *Handler) Handle(ctx context.Context, chatID, name, text string) {
	if err := h.UserRepo.AddUser(ctx, chatID, name); err != nil {
		log.Printf("Ошибка сохранения пользователя %s: %v", chatID, err)
	}

	if draft, ok := h.Drafts[chatID]; ok && draft != nil {
		h.handleEventDraft(ctx, chatID, text, draft)
		return
	}

	switch {
	case text == "/start":
		h.SendWithButtons(chatID,
			"Привет, "+name+"! 👋\n\nВыбери свою роль в компании:",
			[][]botgolang.Button{
				{
					{Text: "👤 Сотрудник", CallbackData: "role:employee"},
					{Text: "🔑 Администратор", CallbackData: "role:admin"},
				},
			},
		)

	case text == "/info":
		role, err := h.UserRepo.GetRole(ctx, chatID)
		if err != nil {
			log.Printf("Ошибка получения роли %s: %v", chatID, err)
			return
		}

		switch role {
		case "employee":
			msg := h.Bot.NewTextMessage(chatID,
				"📖 Доступные команды:\n\n"+
					"/start — начать заново\n"+
					"/info — список команд\n"+
					"/events — список актуальных событий",
			)
			msg.Send()

		case "admin":
			msg := h.Bot.NewTextMessage(chatID,
				"📖 Доступные команды:\n\n"+
					"/start — начать заново\n"+
					"/info — список команд\n"+
					"/event — создать рассылку\n"+
					"/my_events — просмотр всех актуальных событий"+
					"/report — статистика рассылок",
			)
			msg.Send()
		}

	case text == "/event":
		h.createEvent(ctx, chatID)

	case text == "/my_events":
		h.GetAllEvents(ctx, chatID)
	}

}

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

func (h *Handler) HandleCallback(ctx context.Context, chatID, name, data, messageID string) {

	switch data {

	case "role:employee":
		if err := h.UserRepo.UpdateRole(ctx, "employee", chatID); err != nil {
			log.Printf("Ошибка обновления роли %s: %v", chatID, err)
		}

		h.replaceButtons(chatID, messageID, "Выбери свою роль в компании:\n\n👤 Сотрудник выбран ✅")

		h.SendWithButtons(chatID,
			"Отлично! Выбери свой отдел:",
			[][]botgolang.Button{
				{
					{Text: "💻 IT", CallbackData: "dept:it"},
					{Text: "📊 Аналитика", CallbackData: "dept:analytics"},
				},
				{
					{Text: "🎨 Дизайн", CallbackData: "dept:design"},
					{Text: "📖 Менеджмент", CallbackData: "dept:management"},
				},
			},
		)

	case "role:admin":
		if err := h.UserRepo.UpdateRole(ctx, "admin", chatID); err != nil {
			log.Printf("Ошибка обновления роли %s: %v", chatID, err)
		}

		h.replaceButtons(chatID, messageID, "Выбери свою роль в компании:\n\nДобро пожаловать, администратор! 🔑\n\n"+
			"Если появятся вопросы — введите /info")

	case "dept:it":
		h.handleDept(ctx, chatID, messageID, "it", "IT")

	case "dept:analytics":
		h.handleDept(ctx, chatID, messageID, "analytics", "Аналитика")

	case "dept:design":
		h.handleDept(ctx, chatID, messageID, "design", "Дизайн")

	case "dept:management":
		h.handleDept(ctx, chatID, messageID, "management", "Менеджмент")

	case "gender:male":
		h.handleGender(ctx, chatID, messageID, "male", "Мужской")

	case "gender:female":
		h.handleGender(ctx, chatID, messageID, "female", "Женский")

	case "deptEvent:it", "deptEvent:analytics", "deptEvent:design", "deptEvent:management", "deptEvent:all":
		draft, ok := h.Drafts[chatID]
		if !ok || draft == nil || draft.State != "waiting_department" {
			return
		}

		switch data {
		case "deptEvent:it":
			draft.Department = "it"
		case "deptEvent:analytics":
			draft.Department = "analytics"
		case "deptEvent:design":
			draft.Department = "design"
		case "deptEvent:management":
			draft.Department = "management"
		case "deptEvent:all":
			draft.Department = "all"
		}

		draft.State = "waiting_gender"
		h.replaceButtons(chatID, messageID, "Отдел выбран ✅")

		h.SendWithButtons(chatID, "Выберите пол:", [][]botgolang.Button{
			{
				{Text: "Женский", CallbackData: "genderEvent:female"},
				{Text: "Мужской", CallbackData: "genderEvent:male"},
				{Text: "Для всех", CallbackData: "genderEvent:all"},
			},
		})
		return

	case "genderEvent:female", "genderEvent:male", "genderEvent:all":
		draft, ok := h.Drafts[chatID]

		if !ok || draft == nil || draft.State != "waiting_gender" {
			return
		}

		switch data {
		case "genderEvent:female":
			draft.Gender = "female"
		case "genderEvent:male":
			draft.Gender = "male"
		case "genderEvent:all":
			draft.Gender = "all"
		}

		h.replaceButtons(chatID, messageID, "Пол выбран ✅")
		draft.State = "waiting_time_to_send"

		msg := h.Bot.NewTextMessage(chatID, "Введите дату отправки в формате 2006-01-02 15:04")
		if err := msg.Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}

	case "readyEvent:yes", "readyEvent:no":
		draft, ok := h.Drafts[chatID]

		if !ok || draft == nil || draft.State != "ready_to_send" {
			return
		}

		switch data {
		case "readyEvent:yes":

			if err := h.SendEvent(ctx, draft, chatID); err != nil {
				log.Printf("Ошибка сохранения события для %s: %v", chatID, err)

				h.Bot.NewTextMessage(chatID,
					"❌ Не удалось сохранить событие.\n"+
						"Попробуйте позже или обратитесь к разработчику.").Send()
				return
			}
			h.replaceButtons(chatID, messageID, "Событие успешно создано!")
			delete(h.Drafts, chatID)

		case "readyEvent:no":
			delete(h.Drafts, chatID)
			h.replaceButtons(chatID, messageID, "Для создания нового события введите команду: /event")

		}
	}
}
