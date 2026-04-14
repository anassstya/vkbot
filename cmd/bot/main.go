package main

import (
	"context"
	"log"
	"os"

	"github.com/anassstya/vkbot/internal/db"
	"github.com/anassstya/vkbot/internal/handler"
	"github.com/anassstya/vkbot/internal/repository"
	"github.com/anassstya/vkbot/internal/scheduler"
	"github.com/joho/godotenv"
	botgolang "github.com/mail-ru-im/bot-golang"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env:", err)
	}

	botToken := os.Getenv("BOT_TOKEN")
	dbURL := os.Getenv("DATABASE_URL")

	if err := db.RunMigrations(dbURL); err != nil {
		log.Fatal("Migration error:", err)
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer pool.Close()

	bot, err := botgolang.NewBot(botToken)
	if err != nil {
		log.Fatal("wrong token:", err)
	}

	userRepo := repository.NewUserRepo(pool)
	h := handler.NewHandler(userRepo, bot)

	if err := userRepo.EnsureWelcomeTrigger(ctx); err != nil {
		log.Printf("Ошибка создания welcome триггера: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sched := scheduler.NewScheduler(userRepo, bot)
	go sched.Start(ctx)

	updates := bot.GetUpdatesChannel(cancelCtx)

	for update := range updates {
		chatID := update.Payload.From.ID
		name := update.Payload.From.FirstName + " " + update.Payload.From.LastName
		text := update.Payload.Text
		msg := update.Payload.CallbackMessage()

		switch update.Type {
		case botgolang.NEW_MESSAGE:
			h.Handle(ctx, chatID, name, text)

		case botgolang.CALLBACK_QUERY:
			h.HandleCallback(ctx, chatID, name, update.Payload.CallbackData, msg.ID)
		}
	}
}
