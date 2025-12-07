package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"gobotcat/services"
	"gobotcat/storage"
)

type WebhookHandler struct {
	stripe        *services.StripeService
	telegram      *services.TelegramService
	photoStore    storage.PhotoStore
	paymentStore  storage.PaymentStore
}

func NewWebhookHandler(stripe *services.StripeService, telegram *services.TelegramService, photoStore storage.PhotoStore, paymentStore storage.PaymentStore) *WebhookHandler {
	return &WebhookHandler{
		stripe:       stripe,
		telegram:     telegram,
		photoStore:   photoStore,
		paymentStore: paymentStore,
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
		
		sessionID := session["id"].(string)
		userID := session["client_reference_id"].(string)
		paymentStatus := session["payment_status"].(string)
		amountTotal := int64(session["amount_total"].(float64))

		chatID, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			log.Printf("Failed to parse userID: %v", err)
			h.paymentStore.SavePayment(&storage.Payment{
				ID:     sessionID,
				UserID: userID,
				Status: "failed",
			})
			return
		}

		// Сохраняем платёж в БД
		payment := &storage.Payment{
			ID:     sessionID,
			UserID: userID,
			Amount: amountTotal,
			Status: "paid",
		}
		h.paymentStore.SavePayment(payment)

		// Сообщение об успешной оплате
		h.telegram.SendMessage(chatID, "✅ Спасибо! Ваша оплата прошла успешно.")

		if paymentStatus == "paid" {
			
			
			// Получаем список фото владельца
			photo, err := h.photoStore.GetRandomPhoto()
			if err != nil || photo == nil {
				
				h.telegram.SendMessage(chatID, "❌ Фото не найдены")
				h.paymentStore.UpdatePaymentStatus(sessionID, "failed")
				return
			}

			// Отправляем картинку по file_id
			err = h.telegram.SendImageByID(chatID, photo.FileID, "Вот ваша картинка!")
			if err != nil {
				log.Printf("Failed to send image: %v", err)
				h.telegram.SendMessage(chatID, "❌ Ошибка при отправке картинки")
				h.paymentStore.UpdatePaymentStatus(sessionID, "failed")
				return
			}

			// Обновляем статус на успешный
			h.paymentStore.UpdatePaymentStatus(sessionID, "image_sent")
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}
