package main

import (
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gobotcat/config"
	"gobotcat/handlers"
	"gobotcat/services"
	"gobotcat/storage"
)

func main() {
	cfg := config.Load()

	// init database Payments
	db, err := gorm.Open(sqlite.Open("payments.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	paymentStore := storage.NewGormPaymentStore(db)
	
	// init database Photo
	db, err = gorm.Open(sqlite.Open("photo.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	photoStore := storage.NewGormPhotoStore(db)
	

	// init service 
	stripeService := services.NewStripeService(cfg.StripeSecret)
	telegramService, err := services.NewTelegramService(cfg.TelegramKey)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// init handlers
	webhookHandler := handlers.NewWebhookHandler(stripeService, telegramService, photoStore, paymentStore, cfg.StripeWebhookSecret)
	botHandler := handlers.NewBotHandler(telegramService, stripeService, cfg.WebhookURL, photoStore)

	// Маршруты
	http.HandleFunc("/webhook/stripe", webhookHandler.HandleStripeWebhook)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Запуск Telegram bot в горутине
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := telegramService.Bot().GetUpdatesChan(u)
		botHandler.HandleUpdates(updates)
	}()

	// Запуск сервера
	fmt.Printf("Starting server on port %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
