package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CoinbaseService struct {
	apiKey string
}

type CoinbaseCharge struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Address  string `json:"address"`
	Status   string `json:"status"`
}

type CoinbaseChargeRequest struct {
	Name           string `json:"name"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	Metadata       map[string]string `json:"metadata"`
	RedirectURL    string `json:"redirect_url"`
	CancelURL      string `json:"cancel_url"`
	Description    string `json:"description"`
}

type CoinbaseResponse struct {
	Data CoinbaseCharge `json:"data"`
}

func NewCoinbaseService(apiKey string) *CoinbaseService {
	return &CoinbaseService{apiKey: apiKey}
}

// CreateCharge creates a charge (payment request)
func (s *CoinbaseService) CreateCharge(userID string, amount string, currency string, returnURL string) (string, error) {
	payload := CoinbaseChargeRequest{
		Name:     "Image Pack",
		Amount:   amount,
		Currency: currency,
		Metadata: map[string]string{
			"user_id": userID,
		},
		RedirectURL: returnURL + "/payment-success",
		CancelURL:   returnURL + "/payment-canceled",
		Description: "Purchase photo pack",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.commerce.coinbase.com/charges", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CC-Api-Key", s.apiKey)
	req.Header.Set("X-CC-Version", "2018-03-22")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("coinbase api error: %d %s", resp.StatusCode, string(body))
	}

	var result CoinbaseResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Data.HostedURL, nil
}

// ValidateWebhookSignature validates Coinbase webhook signature
func (s *CoinbaseService) ValidateWebhookSignature(body []byte, sig string, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	expectedSig := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expectedSig))
}
