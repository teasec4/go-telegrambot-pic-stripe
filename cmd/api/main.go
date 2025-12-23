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
	"gobotcat/storer"
)

func main() {
	cfg := config.Load()

	// Initialize single database for all data
	db := openDatabase("app.db")
	
	// Initialize storer
	appStorer := storer.NewGormStorer(db)

	// Initialize services
	stripeService := services.NewStripeService(cfg.StripeSecret)
	tronService := services.NewTronService(cfg.TronAPIKey, cfg.TronMainAddress)
	telegramService, err := services.NewTelegramService(cfg.TelegramKey)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(stripeService, telegramService, appStorer, cfg.StripeWebhookSecret)
	tronWebhookHandler := handlers.NewTronWebhookHandler(tronService, telegramService, appStorer)
	botHandler := handlers.NewBotHandler(telegramService, stripeService, tronService, cfg.WebhookURL, appStorer)

	// Parse payment templates
	successTpl := template.Must(template.ParseFiles("templates/success.html"))
	canceledTpl := template.Must(template.ParseFiles("templates/canceled.html"))

	// Routes
	http.HandleFunc("/webhook/stripe", webhookHandler.HandleStripeWebhook)
	http.HandleFunc("/webhook/tron", tronWebhookHandler.HandleTronWebhook)
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

	// Start Tron payment checker in a goroutine
	go tronWebhookHandler.CheckPendingPayments()

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

func openDatabase(dbPath string) *gorm.DB{
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Payments database: %v", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
        log.Fatalf("Failed to get database connection: %v", err)
    }
    defer sqlDB.Close()
    
    return db

}
