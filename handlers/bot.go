package handlers

import (
	"fmt"
	"log"
	"strconv"

	"gobotcat/services"
	"gobotcat/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

		// Handle photos (higher priority than text)
		if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
			if h.telegram.IsAdmin(chatID){
				photo := &storage.Photo{
					// Get the highest quality photo. Telegram provides multiple sizes; the last one has the best quality.
					FileID: update.Message.Photo[len(update.Message.Photo)-1].FileID,
				}
				err := h.handlePhotoUpload(photo)

				// Send a message to the user about the result
				if err != nil {
					h.telegram.SendMessage(chatID, "❌ Failed to save photo")
				} else {
					h.telegram.SendMessage(chatID, "✅ Photo saved successfully!")
				}
			}
			continue
		}

		// Handle text messages
		text := update.Message.Text
		if text == "" {
			continue
		}

		switch text {
		case "/start":
			h.telegram.SendMessage(chatID, "Hello! Type /pay to purchase an image")
		case "/pay":
			h.handlePayment(chatID, userID)
		case "/id":
			fmt.Println(chatID)
		default:
			h.telegram.SendMessage(chatID, "Unknown command. Use /pay")
		}
	}
}

// TODO: implement price selection like 3 price tiers
func (h *BotHandler) handlePayment(chatID int64, userID string) {
	// 9.99 USD for testing
	paymentURL, err := h.stripe.CreatePaymentSession(userID, 999, h.webhookURL)
	if err != nil {
		log.Printf("Failed to create payment session: %v", err)
		h.telegram.SendMessage(chatID, "Failed to create payment session")
		return
	}

	// Create inline button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Pay for photo", paymentURL),
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
