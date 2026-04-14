package handler

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	botgolang "github.com/mail-ru-im/bot-golang"
)

type SessionState string

const (
	stateIdle                    SessionState = ""
	stateWaitingTitle            SessionState = "waiting_title"
	stateWaitingDescription      SessionState = "waiting_description"
	stateWaitingDepartment       SessionState = "waiting_department"
	stateWaitingGender           SessionState = "waiting_gender"
	stateWaitingNotificationType SessionState = "waiting_notification_type"
	stateWaitingTimeToSend       SessionState = "waiting_time_to_send"
	stateWaitingRecurrence       SessionState = "waiting_recurrence"
	stateWaitingRecTime          SessionState = "waiting_rec_time"
	stateWaitingTime             SessionState = "waiting_time"
	stateReadyToSend             SessionState = "ready_to_send"
)

type Session struct {
	ChatID string
	State  SessionState
	Draft  NotificationDraft

	Bot      BotInterface
	UserRepo UserRepoInterface
}

func NewSession(chatID string, bot BotInterface, userRepo UserRepoInterface) *Session {
	return &Session{
		ChatID:   chatID,
		State:    stateIdle,
		Bot:      bot,
		UserRepo: userRepo,
	}
}
func (s *Session) Handle(ctx context.Context, event Event) {
	switch event.Type {
	case "message":
		s.handleMessage(ctx, event)
	case "callback":
		s.handleCallback(ctx, event)
	default:
		log.Printf("неизвестный тип события для chatID=%s: %s", event.ChatID, event.Type)
	}
}

func (s *Session) handleMessage(ctx context.Context, event Event) {
	text := strings.TrimSpace(event.Text)
	if text == "" {
		return
	}

	if s.State != stateIdle {
		if text == "/cancel" {
			s.resetDraft()
			s.sendText(event.ChatID, "Создание уведомления отменено.")
			return
		}

		if strings.HasPrefix(text, "/") {
			s.sendText(event.ChatID, "Сейчас идёт создание уведомления. Сначала завершите его или напишите /cancel.")
			return
		}

		s.handleDraftMessage(ctx, event.ChatID, text)
		return
	}

	switch text {
	case "/start":
		s.resetDraft()

		user, err := s.UserRepo.GetUser(ctx, event.ChatID)
		if err == nil && user.ChatID != "" && user.Role != "" && user.Department != "" && user.Gender != "" {

			roleText := map[string]string{
				"employee": "Сотрудник",
				"admin":    "Администратор",
			}[user.Role]

			s.sendText(event.ChatID, fmt.Sprintf(
				"С возвращением, %s! 👋\n\nВаша роль: %v \nЕсли появятся вопросы — введите /info",
				event.Name,
				roleText,
			))
			return
		}

		s.sendWithButtons(event.ChatID,
			"Привет, "+event.Name+"! 👋\n\nВыбери свою роль в компании:",
			[][]botgolang.Button{
				{
					{Text: "👤 Сотрудник", CallbackData: "role:employee"},
					{Text: "🔑 Администратор", CallbackData: "role:admin"},
				},
			},
		)

	case "/info":
		s.showInfo(ctx, event.ChatID)

	case "/event":
		s.startNotificationFlow(ctx, event.ChatID)

	case "/my_events":
		s.getAllNotifications(ctx, event.ChatID)

	case "/profile":
		s.showProfileData(ctx, event.ChatID)

	case "/change_dept":
		s.showChangeDeptButtons(ctx, event.ChatID)

	case "/report":
		s.showReport(ctx, event.ChatID)
	}
}

func (s *Session) handleDraftMessage(ctx context.Context, chatID, text string) {
	switch s.State {
	case stateWaitingTitle:
		s.Draft.Title = text
		s.State = stateWaitingDescription
		s.sendText(chatID, "Введите описание уведомления:")

	case stateWaitingDescription:
		s.Draft.Description = text
		s.State = stateWaitingDepartment

		s.sendWithButtons(chatID, "Выберите отдел получателей:",
			[][]botgolang.Button{
				{
					{Text: "💻 IT", CallbackData: "deptNotification:it"},
					{Text: "📊 Аналитика", CallbackData: "deptNotification:analytics"},
				},
				{
					{Text: "🎨 Дизайн", CallbackData: "deptNotification:design"},
					{Text: "📖 Менеджмент", CallbackData: "deptNotification:management"},
				},
				{
					{Text: "Все отделы", CallbackData: "deptNotification:all"},
				},
			},
		)

	case stateWaitingTimeToSend, stateWaitingRecTime:
		t, err := time.ParseInLocation("2006-01-02 15:04", text, time.Local)
		if err != nil {
			s.sendText(chatID, "Неверный формат даты. Попробуйте ещё раз: 2006-01-02 15:04")
			return
		}

		if !t.After(time.Now()) {
			s.sendText(chatID, "Дата отправки должна быть позже текущего времени.")
			return
		}

		if s.State == stateWaitingTimeToSend {
			s.Draft.TimeToSend = t
		} else {
			s.Draft.NextSend = t
		}

		s.State = stateReadyToSend
		s.sendPreviewAndConfirm(chatID)

	case stateWaitingTime:
		t, err := time.ParseInLocation("15:04", text, time.Local)
		if err != nil {
			s.sendText(chatID, "Неверный формат времени. Попробуйте ещё раз: 15:04")
			return
		}

		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)

		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
			next = next.Add(24 * time.Hour)
		}

		s.Draft.NextSend = next
		s.State = stateReadyToSend
		s.sendPreviewAndConfirm(chatID)

	default:
		s.sendText(chatID, "Сейчас я не жду текст. Используйте команды или кнопки.")
	}
}

func (s *Session) handleCallback(ctx context.Context, event Event) {
	if strings.HasPrefix(event.Data, "read:") {
		idStr := strings.TrimPrefix(event.Data, "read:")
		notificationID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Ошибка парсинга notificationID: %v", err)
			return
		}

		if err := s.UserRepo.MarkAsRead(ctx, notificationID, event.ChatID); err != nil {
			log.Printf("Ошибка MarkAsRead: %v", err)
		}

		keyboard := botgolang.NewKeyboard()
		keyboard.AddRow(
			botgolang.Button{
				Text: "✅ Прочитано",
			},
		)

		edited := s.Bot.NewInlineKeyboardMessage(event.ChatID, "", keyboard)
		edited.ID = event.MessageID
		_ = edited.Send()

		return
	}

	switch event.Data {
	case "role:employee":
		if err := s.UserRepo.UpdateRole(ctx, "employee", event.ChatID); err != nil {
			log.Printf("Ошибка обновления роли %s: %v", event.ChatID, err)
		}

		s.replaceButtons(event.ChatID, event.MessageID, "Выбери свою роль в компании:\n\n👤 Сотрудник выбран ✅")
		s.sendWithButtons(event.ChatID,
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
		if err := s.UserRepo.UpdateRole(ctx, "admin", event.ChatID); err != nil {
			log.Printf("Ошибка обновления роли %s: %v", event.ChatID, err)
		}
		s.replaceButtons(event.ChatID, event.MessageID, "Добро пожаловать! 🔑\n\nЕсли появятся вопросы — введите /info")

	case "dept:it":
		s.handleDept(ctx, event.ChatID, event.MessageID, "it", "IT")
	case "dept:analytics":
		s.handleDept(ctx, event.ChatID, event.MessageID, "analytics", "Аналитика")
	case "dept:design":
		s.handleDept(ctx, event.ChatID, event.MessageID, "design", "Дизайн")
	case "dept:management":
		s.handleDept(ctx, event.ChatID, event.MessageID, "management", "Менеджмент")

	case "gender:male":
		s.handleGender(ctx, event.ChatID, event.MessageID, "male", "Мужской", event.Name)
	case "gender:female":
		s.handleGender(ctx, event.ChatID, event.MessageID, "female", "Женский", event.Name)

	case "deptNotification:it", "deptNotification:analytics", "deptNotification:design", "deptNotification:management", "deptNotification:all":
		if s.State != stateWaitingDepartment {
			return
		}

		switch event.Data {
		case "deptNotification:it":
			s.Draft.RecipientsDepartment = "it"
		case "deptNotification:analytics":
			s.Draft.RecipientsDepartment = "analytics"
		case "deptNotification:design":
			s.Draft.RecipientsDepartment = "design"
		case "deptNotification:management":
			s.Draft.RecipientsDepartment = "management"
		case "deptNotification:all":
			s.Draft.RecipientsDepartment = "all"
		}

		s.State = stateWaitingGender
		s.replaceButtons(event.ChatID, event.MessageID, "Отдел выбран ✅")

		s.sendWithButtons(event.ChatID, "Выберите пол:", [][]botgolang.Button{
			{
				{Text: "Женский", CallbackData: "genderNotification:female"},
				{Text: "Мужской", CallbackData: "genderNotification:male"},
				{Text: "Для всех", CallbackData: "genderNotification:all"},
			},
		})

	case "genderNotification:female", "genderNotification:male", "genderNotification:all":
		if s.State != stateWaitingGender {
			return
		}

		switch event.Data {
		case "genderNotification:female":
			s.Draft.RecipientsGender = "female"
		case "genderNotification:male":
			s.Draft.RecipientsGender = "male"
		case "genderNotification:all":
			s.Draft.RecipientsGender = "all"
		}

		s.State = stateWaitingNotificationType
		s.replaceButtons(event.ChatID, event.MessageID, "Пол выбран ✅")

		s.sendWithButtons(event.ChatID, "Выберите тип уведомления:", [][]botgolang.Button{
			{
				{Text: "Единоразовое уведомление", CallbackData: "notificationType:single"},
			},
			{
				{Text: "Регулярные уведомления", CallbackData: "notificationType:regular"},
			},
		})

	case "notificationType:single", "notificationType:regular":
		if s.State != stateWaitingNotificationType {
			return
		}

		switch event.Data {
		case "notificationType:single":
			s.Draft.Type = "one_time"
			s.State = stateWaitingTimeToSend

			s.replaceButtons(event.ChatID, event.MessageID, "Тип уведомлений выбран ✅")
			s.sendText(event.ChatID, "Введите дату отправки в формате 2006-01-02 15:04")

		case "notificationType:regular":
			s.Draft.Type = "recurring"
			s.State = stateWaitingRecurrence

			s.replaceButtons(event.ChatID, event.MessageID, "Тип уведомлений выбран ✅")
			s.sendWithButtons(event.ChatID, "Когда отправлять рассылку", [][]botgolang.Button{
				{
					{Text: "Каждый месяц", CallbackData: "recurrence:monthly"},
					{Text: "Каждую неделю", CallbackData: "recurrence:weekly"},
				},
				{
					{Text: "По рабочим дням", CallbackData: "recurrence:workdays"},
				},
			})
		}

	case "recurrence:monthly", "recurrence:weekly", "recurrence:workdays":
		if s.State != stateWaitingRecurrence {
			return
		}

		switch event.Data {
		case "recurrence:monthly":
			s.Draft.Recurrence = "monthly"
			s.State = stateWaitingRecTime

			s.replaceButtons(event.ChatID, event.MessageID, "Ежемесячная рассылка ✅")
			s.sendText(event.ChatID,
				"Введите дату первой отправки в формате 2006-01-02 15:04\n\n"+
					"Эта дата будет использована как дата следующей отправки.\n"+
					"В дальнейшем в указанное время и день рассылка будет выполняться ежемесячно",
			)

		case "recurrence:weekly":
			s.Draft.Recurrence = "weekly"
			s.State = stateWaitingRecTime

			s.replaceButtons(event.ChatID, event.MessageID, "Еженедельная рассылка ✅")
			s.sendText(event.ChatID,
				"Введите дату первой отправки в формате 2006-01-02 15:04\n\n"+
					"Эта дата будет использована как дата следующей отправки.\n"+
					"В дальнейшем в указанное время и день рассылка будет выполняться еженедельно",
			)

		case "recurrence:workdays":
			s.Draft.Recurrence = "workdays"
			s.State = stateWaitingTime

			s.replaceButtons(event.ChatID, event.MessageID, "Рассылка по рабочим дням ✅")
			s.sendText(event.ChatID, "Введите время отправки в формате 15:04\nВ это время каждый рабочий день (пн-пт) будет выполняться рассылка")
		}

	case "readyNotification:yes", "readyNotification:no":
		if s.State != stateReadyToSend {
			return
		}

		switch event.Data {
		case "readyNotification:yes":
			if err := s.SendNotification(ctx, &s.Draft, event.ChatID); err != nil {
				log.Printf("Ошибка сохранения уведомления для %s: %v", event.ChatID, err)
				s.sendText(event.ChatID, "❌ Не удалось сохранить уведомление.\nПопробуйте позже.")
				return
			}

			s.replaceButtons(event.ChatID, event.MessageID, "Уведомление успешно создано!")
			s.resetDraft()

		case "readyNotification:no":
			s.resetDraft()
			s.replaceButtons(event.ChatID, event.MessageID, "Для создания нового уведомления введите команду: /event")
		}
	}
}

func (s *Session) startNotificationFlow(ctx context.Context, chatID string) {
	role, err := s.UserRepo.GetRole(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения роли %s: %v", chatID, err)
		return
	}

	if role != "admin" {
		s.sendText(chatID, "Команда доступна только администратору")
		return
	}

	if s.State != stateIdle {
		s.sendText(chatID, "У вас уже есть незавершённое уведомление. Завершите его или напишите /cancel.")
		return
	}

	s.resetDraft()
	s.State = stateWaitingTitle
	s.sendText(chatID, "Введите название уведомления:")
}

func (s *Session) handleDept(ctx context.Context, chatID, messageID, dept, deptName string) {
	user, err := s.UserRepo.GetUser(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения пользователя %s: %v", chatID, err)
		return
	}

	wasSet := user.Department != ""

	if err := s.UserRepo.AddDept(ctx, dept, chatID); err != nil {
		log.Printf("Ошибка обновления отдела %s: %v", chatID, err)
		return
	}

	if wasSet {
		s.replaceButtons(chatID, messageID, "✅ Отдел изменён на: "+deptName)
		return
	}

	s.replaceButtons(chatID, messageID, "Отдел выбран: "+deptName+" ✅")
	s.sendWithButtons(chatID, "Теперь укажи свой пол:",
		[][]botgolang.Button{
			{
				{Text: "👨 Мужской", CallbackData: "gender:male"},
				{Text: "👩 Женский", CallbackData: "gender:female"},
			},
		},
	)
}

func (s *Session) handleGender(ctx context.Context, chatID, messageID, gender, label, name string) {
	if err := s.UserRepo.AddGender(ctx, gender, chatID); err != nil {
		log.Printf("Ошибка обновления пола %s: %v", chatID, err)
	}

	s.replaceButtons(chatID, messageID, "Пол выбран:\n\n"+label+" ✅")

	notification, err := s.UserRepo.GetWelcomeTrigger(ctx)
	if err != nil {
		log.Printf("Ошибка получения триггера %s: %v", chatID, err)
		return
	}

	title := strings.ReplaceAll(notification.Title, "{name}", name)
	description := strings.ReplaceAll(notification.Description, "{name}", name)

	text := fmt.Sprintf("📢 %s\n\n%s", title, description)

	if err := s.Bot.NewTextMessage(chatID, text).Send(); err != nil {
		log.Printf("Ошибка отправки приветствия %s: %v", chatID, err)
	}
}

func (s *Session) showInfo(ctx context.Context, chatID string) {
	role, err := s.UserRepo.GetRole(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения роли %s: %v", chatID, err)
		return
	}

	switch role {
	case "employee":
		msg := s.Bot.NewTextMessage(chatID,
			"📖 Доступные команды:\n\n"+
				"/start — начать заново\n"+
				"/info — список команд\n"+
				"/profile — просмотр личных данных\n"+
				"/change_dept — изменить отдел",
		)
		if err := msg.Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}

	case "admin":
		msg := s.Bot.NewTextMessage(chatID,
			"📖 Доступные команды:\n\n"+
				"/start — начать заново\n"+
				"/info — список команд\n"+
				"/event — создать рассылку\n"+
				"/my_events — просмотр всех актуальных уведомлений\n"+
				"/report — статистика рассылок",
		)
		if err := msg.Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}

	default:
		msg := s.Bot.NewTextMessage(chatID, "Роль не определена, используйте /start")
		if err := msg.Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}
	}
}

func (s *Session) showChangeDeptButtons(ctx context.Context, chatID string) {
	role, err := s.UserRepo.GetRole(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения роли %s: %v", chatID, err)
		return
	}

	if role != "employee" {
		if err := s.Bot.NewTextMessage(chatID, "Команда доступна только сотрудникам").Send(); err != nil {
			log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
		}
		return
	}

	s.sendWithButtons(chatID, "Выберите новый отдел:",
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
}

func (s *Session) showReport(ctx context.Context, chatID string) {
	role, err := s.UserRepo.GetRole(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения роли %s: %v", chatID, err)
		return
	}

	if role != "admin" {
		s.sendText(chatID, "📊 Доступно только администраторам")
		return
	}

	stats, err := s.UserRepo.GetNotificationStats(ctx)
	if err != nil {
		log.Printf("REPORT ERROR: %v", err)
		s.sendText(chatID, "ошибка")
		return
	}

	var report strings.Builder
	report.WriteString("📊 Статистика рассылок (последние 10):\n\n")

	for i, st := range stats {
		rate := fmt.Sprintf("%.1f%%", st.OpenRate)
		report.WriteString(fmt.Sprintf(
			"%d. %s [ID:%d]\n   📤 %d доставлено | 📖 %d прочитано (%s)\n\n",
			i+1, st.Title, st.ID, st.DeliveredCount, st.ReadCount, rate,
		))
	}

	s.sendText(chatID, report.String())
}

func (s *Session) getAllNotifications(ctx context.Context, chatID string) {
	role, err := s.UserRepo.GetRole(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения роли %s: %v", chatID, err)
		return
	}

	if role != "admin" {
		s.sendText(chatID, "Команда доступна только администратору")
		return
	}

	notifications, err := s.UserRepo.GetMyNotifications(ctx, chatID)
	if err != nil {
		s.sendText(chatID, "Не удалось получить список уведомлений. Попробуйте позже.")
		return
	}

	if len(notifications) == 0 {
		s.sendText(chatID, "📭 Уведомлений пока нет.")
		return
	}

	var builder strings.Builder
	builder.WriteString("📅 Ваши уведомления:\n\n")

	for i, notification := range notifications {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, notification.Title))
		builder.WriteString(fmt.Sprintf("Описание: %s\n", notification.Description))
		builder.WriteString(fmt.Sprintf("Статус: %s\n", notification.Status))
		builder.WriteString(fmt.Sprintf("Тип: %s\n", notification.Type))

		if notification.Type == "one_time" && notification.TimeToSend != nil {
			builder.WriteString(fmt.Sprintf("Дата отправки: %s\n", notification.TimeToSend.Format("2006-01-02 15:04")))
		}

		if notification.Type == "recurring" {
			if notification.Recurrence != nil {
				builder.WriteString(fmt.Sprintf("Повторение: %s\n", *notification.Recurrence))
			}
			if notification.NextSend != nil {
				builder.WriteString(fmt.Sprintf("Следующая отправка: %s\n", notification.NextSend.Format("2006-01-02 15:04")))
			}
		}

		if notification.RecipientsDepartment != "" {
			builder.WriteString(fmt.Sprintf("Отдел: %s\n", notification.RecipientsDepartment))
		}

		if notification.RecipientsGender != "" {
			builder.WriteString(fmt.Sprintf("Пол: %s\n", notification.RecipientsGender))
		}

		builder.WriteString("\n")
	}

	s.sendText(chatID, builder.String())
}

func (s *Session) showProfileData(ctx context.Context, chatID string) {
	user, err := s.UserRepo.GetUser(ctx, chatID)
	if err != nil {
		log.Printf("Ошибка получения пользователя %s: %v", chatID, err)
		return
	}

	dept, gender := translateDeptAndGender(user.Department, user.Gender)

	text := fmt.Sprintf(
		"Имя: %v\nРоль: работник\nПол: %v\nОтдел: %v\nДата присоединения к чат-боту: %v\n",
		user.Name,
		gender,
		dept,
		user.CreatedAt.Format("02.01.2006"),
	)

	s.sendText(chatID, text)
}

func (s *Session) SendNotification(ctx context.Context, draft *NotificationDraft, chatID string) error {
	if draft.Type == "one_time" {
		_, err := s.UserRepo.AddNotification(
			ctx,
			draft.Title,
			draft.Description,
			draft.RecipientsDepartment,
			draft.RecipientsGender,
			chatID,
			draft.TimeToSend,
		)
		return err
	}

	if draft.Type == "recurring" {
		_, err := s.UserRepo.AddRecurringNotification(
			ctx,
			draft.Title,
			draft.Description,
			draft.RecipientsDepartment,
			draft.RecipientsGender,
			chatID,
			draft.Recurrence,
			draft.NextSend,
		)
		return err
	}

	return fmt.Errorf("неизвестный тип уведомления: %s", draft.Type)
}

func (s *Session) resetDraft() {
	s.State = stateIdle
	s.Draft = NotificationDraft{}
}

func (s *Session) sendPreviewAndConfirm(chatID string) {
	dept, gender := translateDeptAndGender(s.Draft.RecipientsDepartment, s.Draft.RecipientsGender)

	var builder strings.Builder
	builder.WriteString("📋 Предпросмотр уведомления:\n\n")
	builder.WriteString("Название: " + s.Draft.Title + "\n")
	builder.WriteString("Описание: " + s.Draft.Description + "\n")
	builder.WriteString("Отдел: " + dept + "\n")
	builder.WriteString("Пол: " + gender + "\n")

	if s.Draft.Type == "one_time" {
		builder.WriteString("Тип: единоразовое\n")
		builder.WriteString("Дата отправки: " + s.Draft.TimeToSend.Format("2006-01-02 15:04") + "\n")
	}

	if s.Draft.Type == "recurring" {
		builder.WriteString("Тип: регулярное\n")
		builder.WriteString("Повторение: " + s.Draft.Recurrence + "\n")
		builder.WriteString("Следующая отправка: " + s.Draft.NextSend.Format("2006-01-02 15:04") + "\n")
	}

	s.sendText(chatID, builder.String())

	s.sendWithButtons(chatID, "Подтвердить создание?", [][]botgolang.Button{
		{
			{Text: "✅ Создать", CallbackData: "readyNotification:yes"},
			{Text: "❌ Отмена", CallbackData: "readyNotification:no"},
		},
	})
}

func (s *Session) sendText(chatID, text string) {
	msg := s.Bot.NewTextMessage(chatID, text)
	if err := msg.Send(); err != nil {
		log.Printf("Ошибка отправки сообщения %s: %v", chatID, err)
	}
}

func (s *Session) sendWithButtons(chatID, text string, buttons [][]botgolang.Button) {
	keyboard := botgolang.NewKeyboard()

	for _, row := range buttons {
		keyboard.AddRow(row...)
	}

	msg := s.Bot.NewInlineKeyboardMessage(chatID, text, keyboard)
	if err := msg.Send(); err != nil {
		log.Printf("Ошибка отправки с кнопками %s: %v", chatID, err)
	}
}

func (s *Session) replaceButtons(chatID, messageID, text string) {
	edited := s.Bot.NewTextMessage(chatID, text)
	edited.ID = messageID

	if err := edited.Edit(); err != nil {
		log.Printf("Ошибка редактирования сообщения: %v", err)
	}
}
