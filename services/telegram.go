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

// IsAdmin checks if a user is an administrator
// TODO: Move admin IDs to configuration instead of hardcoding
// For now, update this slice with your admin IDs
func (t *TelegramService) IsAdmin(chatID int64) bool {
	admins := []int64{5147599417} // Update with actual admin IDs
	for _, adminID := range admins {
		if adminID == chatID {
			return true
		}
	}
	return false
}
