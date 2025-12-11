package services

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramService struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramService(token string) (*TelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramService{bot: bot}, nil
}

// SendImage sends an image by URL
func (t *TelegramService) SendImage(chatID int64, imageURL string, caption string) error {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imageURL))
	photo.Caption = caption

	_, err := t.bot.Send(photo)
	return err
}

// SendMessage sends a text message
func (t *TelegramService) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := t.bot.Send(msg)
	return err
}


// Bot returns the bot instance for direct access
func (t *TelegramService) Bot() *tgbotapi.BotAPI {
	return t.bot
}

// Admin represents an administrator for photo uploads. Not the best practice but simple and works for now.
type Admin struct{
	ChatID int64
}

// Manually add an admin's ID. You can get it in your terminal by typing /id in the bot.
var admins = []Admin{Admin{ChatID: 5147599417}}

func (t *TelegramService) IsAdmin(chatID int64) bool {
	for _, admin := range admins {
		if admin.ChatID == chatID {
			return true
		}
	}
	return false
}
