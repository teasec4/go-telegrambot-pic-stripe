package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	StripeKey           string
	StripeSecret        string
	StripeWebhookSecret string
	TelegramKey         string
	WebhookURL          string
	Port                string
	CoinbaseAPIKey      string
	TronAPIKey          string
	TronMainAddress     string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		StripeKey:           getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StripeSecret:        getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		TelegramKey:         getEnv("TELEGRAM_BOT_TOKEN", ""),
		WebhookURL:          getEnv("WEBHOOK_URL", "http://localhost:8080"),
		Port:                getEnv("PORT", "8080"),
		CoinbaseAPIKey:      getEnv("COINBASE_API_KEY", ""),
		TronAPIKey:          getEnv("TRON_API_KEY", ""),
		TronMainAddress:     getEnv("TRON_MAIN_ADDRESS", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
