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

	// Initialize database and storer
	db := openDatabase("app.db")
	appStorer := storer.NewGormStorer(db)

	// Initialize services
	svc := services.NewServicesFromConfig(cfg)

	// Initialize handlers
	h := handlers.NewHandlers(svc, appStorer, cfg.WebhookURL, cfg.StripeWebhookSecret)

	// Parse payment templates
	successTpl := template.Must(template.ParseFiles("templates/success.html"))
	canceledTpl := template.Must(template.ParseFiles("templates/canceled.html"))

	// Routes
	http.HandleFunc("/webhook/stripe", h.Webhook.HandleStripeWebhook)
	http.HandleFunc("/webhook/tron", h.TronWebhook.HandleTronWebhook)
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
	go h.TronWebhook.CheckPendingPayments()

	// Start Telegram bot in a goroutine
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := svc.Telegram.Bot().GetUpdatesChan(u)
		h.Bot.HandleUpdates(updates)
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
