package handlers

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"gobotcat/services"
	"gobotcat/storage"
)

type WebhookHandler struct {
	stripe   *services.StripeService
	telegram *services.TelegramService
	storage  storage.Storage
}

func NewWebhookHandler(stripe *services.StripeService, telegram *services.TelegramService, store storage.Storage) *WebhookHandler {
	return &WebhookHandler{
		stripe:   stripe,
		telegram: telegram,
		storage:  store,
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

			// TODO: получить ownerID из session, пока используем статический ID для тестирования
			ownerID := "5147599417" // ID владельца фото
			
			// Получаем список фото владельца
			photos, err := h.storage.GetPhotos(ownerID)
			if err != nil || len(photos) == 0 {
				log.Printf("No photos found for owner: %s", ownerID)
				h.telegram.SendMessage(chatID, "❌ Фото не найдены")
				return
			}

			// Выбираем рандомное фото
			randomIndex := rand.Intn(len(photos))
			fileID := photos[randomIndex]

			// Отправляем картинку по file_id
			err = h.telegram.SendImageByID(chatID, fileID, "Вот ваша картинка!")
			if err != nil {
				log.Printf("Failed to send image: %v", err)
				h.telegram.SendMessage(chatID, "❌ Ошибка при отправке картинки")
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}
