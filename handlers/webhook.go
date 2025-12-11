package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
	"gobotcat/services"
	"gobotcat/storage"
)

type WebhookHandler struct {
	stripe             *services.StripeService
	telegram           *services.TelegramService
	photoStore         storage.PhotoStore
	paymentStore       storage.PaymentStore
	webhookSecret      string
}

func NewWebhookHandler(stripe *services.StripeService, telegram *services.TelegramService, photoStore storage.PhotoStore, paymentStore storage.PaymentStore, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		stripe:        stripe,
		telegram:      telegram,
		photoStore:    photoStore,
		paymentStore:  paymentStore,
		webhookSecret: webhookSecret,
	}
}

func (h *WebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	sig := r.Header.Get("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(body, sig, h.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("⚠️  Webhook signature verification failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		// making a struct data for woking on late (sess)
		err := json.Unmarshal(event.Data.Raw, &sess)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.handleCheckoutSessionCompleted(sess)

	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

func (h *WebhookHandler) handleCheckoutSessionCompleted(sess stripe.CheckoutSession) {
	// userID and chatID actualy the same, in telegram chat bot see you like a chatID, but for Stripe make more sence be use "userID" (chatID), or I make it wrong sorry))
	userID := sess.ClientReferenceID //string
	chatID, err := strconv.ParseInt(userID, 10, 64) // same but int
	if err != nil {
		log.Printf("Failed to parse userID: %v\n", err)
		return
	}

	payment := &storage.Payment{
		ID:     sess.ID,
		UserID: userID,
		Amount: sess.AmountTotal,
		Status: "paid",
	}
	h.paymentStore.SavePayment(payment)

	h.telegram.SendMessage(chatID, "✅ Thank you! Your payment was successful.")

	if sess.PaymentStatus == "paid" {
		photo, err := h.photoStore.GetRandomPhoto()
		if err != nil || photo == nil {
			h.telegram.SendMessage(chatID, "❌ No photos found")
			h.paymentStore.UpdatePaymentStatus(sess.ID, "failed")
			return
		}

		err = h.telegram.SendImage(chatID, photo.FileID, "Here is your image!")
		if err != nil {
			log.Printf("Failed to send image: %v\n", err)
			h.telegram.SendMessage(chatID, "❌ Error sending image")
			h.paymentStore.UpdatePaymentStatus(sess.ID, "failed")
			return
		}

		h.paymentStore.UpdatePaymentStatus(sess.ID, "image_sent")
	}
}
