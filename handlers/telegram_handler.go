package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"gobotcat/services"
	"gobotcat/storer"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	services   *services.Services
	webhookURL string
	storer     *storer.GormStorer
}

func NewBotHandler(svc *services.Services, webhookURL string, storer *storer.GormStorer) *BotHandler {
	return &BotHandler{
		services:   svc,
		webhookURL: webhookURL,
		storer:     storer,
	}
}

func (h *BotHandler) HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		// Handle callback queries (button clicks)
		if update.CallbackQuery != nil {
			h.handleCallback(update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userID := strconv.FormatInt(update.Message.From.ID, 10)

		// Handle photos (higher priority than text)
		if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
			if h.services.Telegram.IsAdmin(chatID){
				photo := &storer.Photo{
					// Get the highest quality photo. Telegram provides multiple sizes; the last one has the best quality.
					FileID: update.Message.Photo[len(update.Message.Photo)-1].FileID,
				}
				err := h.handlePhotoUpload(photo)

				// Send a message to the user about the result
				if err != nil {
					h.services.Telegram.SendMessage(chatID, "❌ Failed to save photo")
					} else {
					h.services.Telegram.SendMessage(chatID, "✅ Photo saved successfully!")
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
			h.handleStart(chatID)
		case "/pay":
			h.handlePaymentMenu(chatID, userID)
		case "/id":
			fmt.Println(chatID)
		default:
			h.services.Telegram.SendMessage(chatID, "Unknown command. Use /pay")
		}
	}
}

func (h *BotHandler) handleStart(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Pay with Stripe", "pay_stripe"),
			tgbotapi.NewInlineKeyboardButtonData("Pay with USDT", "pay_usdt"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Welcome! Choose payment method to buy photos")
	msg.ReplyMarkup = keyboard
	h.services.Telegram.Bot().Send(msg)
}

func (h *BotHandler) handlePaymentMenu(chatID int64, userID string) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Stripe ($9.99)", "pay_stripe"),
			tgbotapi.NewInlineKeyboardButtonData("USDT (10)", "pay_usdt"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Select payment method:")
	msg.ReplyMarkup = keyboard
	h.services.Telegram.Bot().Send(msg)
}

// Handle Stripe payment
func (h *BotHandler) handleStripePayment(chatID int64, userID string) {
	h.services.Telegram.SendMessage(chatID, "⏳ Preparing your payment link... Please wait a moment")

	paymentURL, err := h.services.Stripe.CreatePaymentSession(userID, 999, h.webhookURL)
	if err != nil {
		log.Printf("Failed to create payment session: %v", err)
		h.services.Telegram.SendMessage(chatID, "❌ Failed to create payment session")
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Pay $9.99", paymentURL),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Click the button to pay via Stripe")
	msg.ReplyMarkup = keyboard
	h.services.Telegram.Bot().Send(msg)
}

// Handle USDT payment
func (h *BotHandler) handleUSDTPayment(chatID int64, userID string) {
	h.services.Telegram.SendMessage(chatID, "⏳ Preparing your payment address... Please wait a moment")

	mainAddress := h.services.Tron.GetMainAddress()
	if mainAddress == "" {
		log.Printf("ERROR: Tron main address not configured in TRON_MAIN_ADDRESS env var")
		h.services.Telegram.SendMessage(chatID, "❌ USDT payment is not configured. Contact admin.")
		return
	}

	payment := h.createOrUpdateTronPayment(userID, mainAddress)
	if payment == nil {
		h.services.Telegram.SendMessage(chatID, "Failed to create payment")
		return
	}

	// Send payment address to user
	tokenName := "TRX"
	tokenAmount := "10"
	message := "Send exactly " + tokenAmount + " " + tokenName + " to this Tron address:\n\n" +
		"`" + mainAddress + "`\n\n" +
		"Network: Tron (Shasta Testnet)\n" +
		"Amount: " + tokenAmount + " " + tokenName + "\n" +
		"Expires in: 24 hours\n\n" +
		"We'll notify you when payment is confirmed."

	h.services.Telegram.SendMessage(chatID, message)
}

func (h *BotHandler) handleCallback(query *tgbotapi.CallbackQuery) {
	chatID := query.From.ID
	userID := strconv.FormatInt(query.From.ID, 10)
	
	// Answer the callback to remove the loading state
	callback := tgbotapi.NewCallback(query.ID, "")
	h.services.Telegram.Bot().Request(callback)

	// Handle different callback data
	switch query.Data {
	case "pay_stripe":
		h.handleStripePayment(chatID, userID)
	case "pay_usdt":
		h.handleUSDTPayment(chatID, userID)
	default:
		h.services.Telegram.SendMessage(chatID, "Unknown action")
	}
}

func (h *BotHandler) handlePhotoUpload(photo *storer.Photo) error{
	err := h.storer.SavePhoto(photo)
	if err != nil {
		log.Printf("Failed to save photo: %v", err)
		return err
	}

	return nil
}

func (h *BotHandler) createOrUpdateTronPayment(userID, address string) *storer.Payment {
	existingPayment, err := h.storer.GetTronPaymentByAddress(address)
	
	if err == nil && existingPayment != nil {
		existingPayment.UserID = userID
		existingPayment.Amount = 10_000_000 // 10 TRX with 6 decimals
		existingPayment.AmountUSD = 10.0
		existingPayment.Status = "pending"
		existingPayment.TxID = ""
		existingPayment.CreatedAt = time.Now()
		existingPayment.ExpiresAt = time.Now().Unix() + 86400 // 24 hours
		
		if err := h.storer.UpdateTronPayment(existingPayment); err != nil {
			log.Printf("Failed to update payment: %v", err)
			return nil
		}
		return existingPayment
	}
	
	payment := &storer.Payment{
		UserID:    userID,
		Type:      "tron",
		Address:   address,
		Amount:    10_000_000, // 10 TRX with 6 decimals
		AmountUSD: 10.0,
		Status:    "pending",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Unix() + 86400, // 24 hours
	}

	if err := h.storer.SaveTronPayment(payment); err != nil {
		log.Printf("Failed to save payment: %v", err)
		return nil
	}
	
	return payment
}
