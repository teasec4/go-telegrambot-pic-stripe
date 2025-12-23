package storer

import "time"

type Photo struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	FileID   string `json:"file_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"index" json:"user_id"`
	Type          string    `json:"type"` // "stripe" or "tron"
	Amount        int64     `json:"amount"`
	AmountUSD     float64   `json:"amount_usd,omitempty"` // для tron
	Status        string    `json:"status"` // "pending", "paid", "confirmed", "image_sent", "failed"
	Error         string    `json:"error,omitempty"`
	Address       string    `json:"address,omitempty"` // для tron платежей
	TxID          string    `gorm:"index" json:"tx_id,omitempty"` // для tron платежей
	Confirmations int64     `json:"confirmations,omitempty"` // для tron
	BlockNumber   int64     `json:"block_number,omitempty"` // для tron
	ExpiresAt     int64     `json:"expires_at,omitempty"` // для tron
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ConfirmedAt   time.Time `json:"confirmed_at,omitempty"`
}
