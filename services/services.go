package services

import (
	"log"

	"gobotcat/config"
)

type Services struct {
	Stripe   *StripeService
	Tron     *TronService
	Telegram *TelegramService
}

func NewServicesFromConfig(cfg *config.Config) *Services {
	stripeService := NewStripeService(cfg.StripeSecret)
	tronService := NewTronService(cfg.TronAPIKey, cfg.TronMainAddress)
	telegramService, err := NewTelegramService(cfg.TelegramKey)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	return &Services{
		Stripe:   stripeService,
		Tron:     tronService,
		Telegram: telegramService,
	}
}
