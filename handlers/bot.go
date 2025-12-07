package handlers

import (
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gobotcat/services"
	"gobotcat/storage"
)

type BotHandler struct {
	telegram *services.TelegramService
	stripe   *services.StripeService
	webhookURL string
	photoStore storage.PhotoStore
}

func NewBotHandler(telegram *services.TelegramService, stripe *services.StripeService, webhookURL string, photoStore storage.PhotoStore) *BotHandler {
	return &BotHandler{
		telegram: telegram,
		stripe: stripe,
		webhookURL: webhookURL,
		photoStore: photoStore,
	}
}

func (h *BotHandler) HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userID := strconv.FormatInt(update.Message.From.ID, 10)

		// Обработка фото (приоритет выше чем текст)
		if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
			if h.telegram.IsAdmin(chatID){
				// передать сюда Photo
				photo := &storage.Photo{
					FileID: update.Message.Photo[len(update.Message.Photo)-1].FileID,
				}
				err := h.handlePhotoUpload(photo)

				// и сообщение юзеру что успешно
				if err != nil {
					h.telegram.SendMessage(chatID, "❌ Ошибка при сохранении фото")
				} else {
					h.telegram.SendMessage(chatID, "✅ Фото сохранено!")
				}
			}
			continue
		}

		// Обработка текста
		text := update.Message.Text
		if text == "" {
			continue
		}

		switch text {
		case "/start":
			h.telegram.SendMessage(chatID, "Привет! Напиши /pay для оплаты картинки")
		case "/pay":
			h.handlePayment(chatID, userID)
		case "/id":
			fmt.Println(chatID)
		default:
			h.telegram.SendMessage(chatID, "Неизвестная команда. Используй /pay")
		}
	}
}

// need to implement adding amount like 3 price range
func (h *BotHandler) handlePayment(chatID int64, userID string) {
	// 9.99 USD for testing
	paymentURL, err := h.stripe.CreatePaymentSession(userID, 999, h.webhookURL)
	if err != nil {
		log.Printf("Failed to create payment session: %v", err)
		h.telegram.SendMessage(chatID, "Ошибка при создании платежа")
		return
	}

	// creating InlineButton
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Оплатить", paymentURL),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Here is a link to pay by Stripe")
	msg.ReplyMarkup = keyboard
	h.telegram.Bot().Send(msg)
}

func (h *BotHandler) handlePhotoUpload(photo *storage.Photo) error{
	err := h.photoStore.SavePhoto(photo)
	if err != nil {
		log.Printf("Failed to save photo: %v", err)
		return err
	}

	return nil
}
