package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	StripeKey    string
	StripeSecret string
	TelegramKey  string
	WebhookURL   string
	Port         string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		StripeKey:    getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StripeSecret: getEnv("STRIPE_SECRET_KEY", ""),
		TelegramKey:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		WebhookURL:   getEnv("WEBHOOK_URL", "http://localhost:8080"),
		Port:         getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
