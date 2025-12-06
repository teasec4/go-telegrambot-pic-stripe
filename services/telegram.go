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

// SendImage отправляет картинку по URL
func (t *TelegramService) SendImage(chatID int64, imageURL string, caption string) error {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imageURL))
	photo.Caption = caption

	_, err := t.bot.Send(photo)
	return err
}

// SendImageByID отправляет картинку по file_id
func (t *TelegramService) SendImageByID(chatID int64, fileID string, caption string) error {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
	photo.Caption = caption

	_, err := t.bot.Send(photo)
	return err
}

// SendMessage отправляет текстовое сообщение
func (t *TelegramService) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := t.bot.Send(msg)
	return err
}


// Bot возвращает экземпляр бота для прямого доступа
func (t *TelegramService) Bot() *tgbotapi.BotAPI {
	return t.bot
}

type Admin struct{
	ChatID int64
}

var admins = []Admin{Admin{ChatID: 5147599417}}

func (t *TelegramService) IsAdmin(chatID int64) bool {
	for _, admin := range admins {
		if admin.ChatID == chatID {
			return true
		}
	}
	return false
}
