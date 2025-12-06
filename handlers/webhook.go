package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"gobotcat/services"
)

type WebhookHandler struct {
	stripe   *services.StripeService
	telegram *services.TelegramService
}

func NewWebhookHandler(stripe *services.StripeService, telegram *services.TelegramService) *WebhookHandler {
	return &WebhookHandler{
		stripe:   stripe,
		telegram: telegram,
	}
}

func (h *WebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Проверяем подпись (нужен endpointSecret из переменных окружения)
	// event, err := h.stripe.ValidateWebhookSignature(body, sig, endpointSecret)

	var event map[string]interface{}
	err = json.Unmarshal(body, &event)
	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	// Обработка события платежа
	if event["type"] == "checkout.session.completed" {
		data := event["data"].(map[string]interface{})
		session := data["object"].(map[string]interface{})
		
		userID := session["client_reference_id"].(string)
		paymentStatus := session["payment_status"].(string)

		if paymentStatus == "paid" {
			chatID, err := strconv.ParseInt(userID, 10, 64)
			if err != nil {
				log.Printf("Failed to parse userID: %v", err)
				http.Error(w, "invalid user id", http.StatusBadRequest)
				return
			}

			// Сообщение об успешной оплате
			h.telegram.SendMessage(chatID, "✅ Спасибо! Ваша оплата прошла успешно.")

			// Отправляем картинку
			imageURL := "https://via.placeholder.com/300?text=Your+Image"
			err = h.telegram.SendImage(chatID, imageURL, "Вот ваша картинка!")
			if err != nil {
				log.Printf("Failed to send image: %v", err)
				http.Error(w, "failed to send image", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}
