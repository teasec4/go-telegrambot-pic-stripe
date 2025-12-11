package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"gobotcat/services"
	"gobotcat/storage"
)

// TronWebhookHandler processes incoming Tron blockchain events and payment confirmations
// Coordinates between Tron blockchain, payment storage, and Telegram notifications
type TronWebhookHandler struct {
	tronService   *services.TronService     // Service for Tron blockchain interactions
	telegramSvc   *services.TelegramService // Service for sending Telegram messages to users
	tronPayments  storage.TronPaymentStore  // Database store for payment records
	photoStore    storage.PhotoStore        // Database store for photo data
}

// TronWebhookPayload represents a payment notification received from Tron webhook or polling
type TronWebhookPayload struct {
	TxID      string `json:"txID"`        // Transaction ID on blockchain
	From      string `json:"from"`        // Sender's Tron address
	To        string `json:"to"`          // Recipient's Tron address (our payment address)
	Amount    int64  `json:"amount"`      // Transaction amount in smallest units (sun)
	Confirmed bool   `json:"confirmed"`   // Whether transaction has sufficient confirmations
	BlockNum  int64  `json:"blockNumber"` // Block number containing the transaction
}

// NewTronWebhookHandler creates a new handler for Tron payment events
func NewTronWebhookHandler(
	tronSvc *services.TronService,
	telegramSvc *services.TelegramService,
	tronPayments storage.TronPaymentStore,
	photoStore storage.PhotoStore,
) *TronWebhookHandler {
	return &TronWebhookHandler{
		tronService:  tronSvc,
		telegramSvc:  telegramSvc,
		tronPayments: tronPayments,
		photoStore:   photoStore,
	}
}

// HandleTronWebhook processes incoming Tron transaction webhook notifications
// NOTE: This is currently not active - we use polling (CheckPendingPayments) instead
// To use this in production: set up a webhook service like Blockhive or use a blockchain listener
func (h *TronWebhookHandler) HandleTronWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload TronWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload"})
		return
	}

	// Get payment record by recipient address
	payment, err := h.tronPayments.GetTronPaymentByAddress(payload.To)
	if err != nil {
		w.WriteHeader(http.StatusOK) // Still return 200 to not retry webhook
		return
	}

	// Convert payment amount from smallest units to display units (6 decimals for TRX/USDT)
	amountDisplay := float64(payload.Amount) / 1e6
	
	// Update payment record with transaction details from webhook
	payment.TxID = payload.TxID
	payment.Amount = payload.Amount
	payment.BlockNumber = payload.BlockNum
	payment.Status = "confirmed"
	payment.ConfirmedAt = time.Now().Unix()
	payment.Confirmations = 25 // Assume webhook indicates sufficient confirmations

	if err := h.tronPayments.UpdateTronPayment(payment); err != nil {
		log.Printf("Failed to update payment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send notification to user via Telegram
	userID, _ := strconv.ParseInt(payment.UserID, 10, 64)
	h.telegramSvc.SendMessage(userID, 
		"✅ Payment received! "+
		"Amount: "+strconv.FormatFloat(amountDisplay, 'f', 2, 64)+" USDT\n"+
		"TxID: "+payload.TxID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// CheckPendingPayments continuously polls for payment confirmations on pending addresses
// TESTNET MODE: Uses polling every 30 seconds (simple approach for testing)
// NOTE: This is a polling approach suitable for testnet only. For production:
// - TODO: Replace with webhooks or blockchain event subscriptions for better efficiency
// - TODO: Increase polling interval to 5-10 minutes to reduce API calls on mainnet
// - TODO: Add exponential backoff for failed checks
// - TODO: Implement state machine for payment status transitions
// - TODO: Add proper error logging and monitoring for stuck payments
func (h *TronWebhookHandler) CheckPendingPayments() {
	// TODO: Make polling interval configurable (currently 30 seconds for testing)
	// Note: For mainnet, increase to 300-600 seconds to avoid excessive API usage
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("[TRON] Checking pending payments...")
		
		// Fetch all pending payments from database
		payments, err := h.tronPayments.GetPendingTronPayments()
		if err != nil {
			log.Printf("[TRON] Failed to get pending payments: %v", err)
			continue
		}

		log.Printf("[TRON] Found %d pending payments", len(payments))

		for _, payment := range payments {
			log.Printf("[TRON] Checking payment for address %s, expected amount: %d sun", payment.Address, payment.Amount)
			
			// Check if address expired (24-hour TTL for payment addresses)
			if time.Now().Unix()-payment.CreatedAt > 86400 {
				log.Printf("[TRON] Payment expired for address %s", payment.Address)
				payment.Status = "expired"
				h.tronPayments.UpdateTronPayment(&payment)
				continue
			}

			// Query Tron blockchain for current balance on payment address
			balance, err := h.tronService.CheckBalance(payment.Address)
			if err != nil {
				log.Printf("[TRON] Failed to check balance for %s: %v", payment.Address, err)
				continue
			}

			log.Printf("[TRON] Balance for %s: %d sun (expected: %d)", payment.Address, balance.Amount, payment.Amount)

			// If received amount matches or exceeds expected amount, mark as confirmed
			if balance.Amount >= payment.Amount && payment.Amount > 0 {
				log.Printf("[TRON] Payment received! Processing confirmation...")
				
				// TESTNET MODE: Simple confirmation by balance check
				// NOTE: This simple confirmation by balance works for testing but has limitations:
				// - Cannot distinguish between different incoming transactions on the same address
				// - Cannot verify exact transaction amount if multiple payments received
				// - No on-chain confirmation verification
				// TODO: For mainnet, implement proper transaction verification with:
				// - Real TxID tracking from blockchain transaction details
				// - Multiple block confirmation requirements
				// - Exact transaction amount validation with hash verification
				
				// Generate a synthetic TxID for testnet (in production, get from blockchain)
				if payment.TxID == "" {
					payment.TxID = "testnet-" + strconv.FormatInt(time.Now().Unix(), 10)
				}
				
				payment.Status = "confirmed"
				payment.ConfirmedAt = time.Now().Unix()
				payment.Confirmations = 25 // Hardcoded for testing, should check actual confirmations on mainnet

				if err := h.tronPayments.UpdateTronPayment(&payment); err != nil {
					log.Printf("[TRON] Failed to update payment: %v", err)
					continue
				}

				log.Printf("[TRON] Payment confirmed! TxID: %s", payment.TxID)

				// Notify user via Telegram that payment was received and confirmed
				userID, _ := strconv.ParseInt(payment.UserID, 10, 64)
				amount := float64(payment.Amount) / 1e6
				h.telegramSvc.SendMessage(userID,
					"✅ Payment confirmed!\n"+
					"Amount: "+strconv.FormatFloat(amount, 'f', 2, 64)+" TRX\n"+
					"TxID: "+payment.TxID)

				// Get random photo from database and send to user
				photo, err := h.photoStore.GetRandomPhoto()
				if err != nil {
					log.Printf("[TRON] Failed to get random photo: %v", err)
					continue
				}

				if photo == nil {
					log.Printf("[TRON] No photos available in database for user %s", payment.UserID)
					h.telegramSvc.SendMessage(userID, "Sorry, no photos available right now.")
					continue
				}

				// Send photo to user
				err = h.telegramSvc.SendImage(userID, photo.FileID, "Your reward photo for the payment!")
				if err != nil {
					log.Printf("[TRON] Failed to send photo to user %s: %v", payment.UserID, err)
					continue
				}

				log.Printf("[TRON] Photo sent successfully to user %s", payment.UserID)
			}
		}
	}
}
