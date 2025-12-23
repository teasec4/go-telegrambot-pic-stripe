package handlers

import (
	"gobotcat/services"
	"gobotcat/storer"
)

type Handlers struct {
	Webhook     *WebhookHandler
	TronWebhook *TronWebhookHandler
	Bot         *BotHandler
}

func NewHandlers(svc *services.Services, appStorer *storer.GormStorer, webhookURL string, webhookSecret string) *Handlers {
	return &Handlers{
		Webhook:     NewWebhookHandler(svc, appStorer, webhookSecret),
		TronWebhook: NewTronWebhookHandler(svc, appStorer),
		Bot:         NewBotHandler(svc, webhookURL, appStorer),
	}
}
