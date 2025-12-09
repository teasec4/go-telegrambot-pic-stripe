package main

import (
	"fmt"
	"html/template"
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

	// Initialize Payments database
	db, err := gorm.Open(sqlite.Open("payments.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Payments database: %v", err)
	}
	paymentStore := storage.NewGormPaymentStore(db)
	
	// Initialize Photo database
	db, err = gorm.Open(sqlite.Open("photo.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Photo database: %v", err)
	}
	photoStore := storage.NewGormPhotoStore(db)
	

	// Initialize services
	stripeService := services.NewStripeService(cfg.StripeSecret)
	telegramService, err := services.NewTelegramService(cfg.TelegramKey)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(stripeService, telegramService, photoStore, paymentStore, cfg.StripeWebhookSecret)
	botHandler := handlers.NewBotHandler(telegramService, stripeService, cfg.WebhookURL, photoStore)

	// Parse payment templates
	successTpl := template.Must(template.ParseFiles("templates/success.html"))
	canceledTpl := template.Must(template.ParseFiles("templates/canceled.html"))

	// Routes
	http.HandleFunc("/webhook/stripe", webhookHandler.HandleStripeWebhook)
	http.HandleFunc("/payment-success", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		successTpl.Execute(w, nil)
	})
	http.HandleFunc("/payment-canceled", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		canceledTpl.Execute(w, nil)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Start Telegram bot in a goroutine
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := telegramService.Bot().GetUpdatesChan(u)
		botHandler.HandleUpdates(updates)
	}()

	// Start server
	fmt.Printf("Starting server on port %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
