package storage

import "gorm.io/gorm"

// TronPayment represents a Tron blockchain payment record
// Stores payment details, transaction info, and confirmation status
type TronPayment struct {
	ID              uint   `gorm:"primaryKey"`            // Primary key
	UserID          string `gorm:"index"`                 // Telegram user ID
	Address         string `gorm:"uniqueIndex"`           // Generated Tron address for receiving payment
	TxID            string `gorm:"index"`                 // Transaction ID on blockchain
	Amount          int64  // Payment amount in native units (with decimals: 6 for TRX/USDT)
	AmountUSD       float64                               // Amount converted to USD
	Status          string                                // Payment status: "pending", "confirmed", "failed", "expired"
	Confirmations   int64                                 // Number of blockchain confirmations
	BlockNumber     int64                                 // Block number where transaction was confirmed
	CreatedAt       int64                                 // Timestamp when payment record was created
	ConfirmedAt     int64                                 // Timestamp when payment was confirmed
	ExpiresAt       int64                                 // Payment address expires after 24 hours (TTL)
}

// TronPaymentStore defines the interface for Tron payment persistence
type TronPaymentStore interface {
	SaveTronPayment(payment *TronPayment) error
	GetTronPayment(txID string) (*TronPayment, error)
	GetTronPaymentByAddress(address string) (*TronPayment, error)
	GetTronPaymentsByUserID(userID string) ([]TronPayment, error)
	UpdateTronPayment(payment *TronPayment) error
	GetPendingTronPayments() ([]TronPayment, error)
}

// GormTronPaymentStore implements TronPaymentStore using GORM for database persistence
type GormTronPaymentStore struct {
	db *gorm.DB
}

// NewGormTronPaymentStore creates a new GORM-based Tron payment store and auto-migrates the schema
func NewGormTronPaymentStore(db *gorm.DB) *GormTronPaymentStore {
	db.AutoMigrate(&TronPayment{})
	return &GormTronPaymentStore{db: db}
}

// SaveTronPayment creates a new payment record in the database
func (s *GormTronPaymentStore) SaveTronPayment(payment *TronPayment) error {
	return s.db.Create(payment).Error
}

// GetTronPayment retrieves a payment record by transaction ID
func (s *GormTronPaymentStore) GetTronPayment(txID string) (*TronPayment, error) {
	var payment TronPayment
	err := s.db.Where("tx_id = ?", txID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetTronPaymentByAddress retrieves a payment record by the unique receiving address
func (s *GormTronPaymentStore) GetTronPaymentByAddress(address string) (*TronPayment, error) {
	var payment TronPayment
	err := s.db.Where("address = ?", address).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetTronPaymentsByUserID retrieves all payment records for a user, ordered by creation time (newest first)
func (s *GormTronPaymentStore) GetTronPaymentsByUserID(userID string) ([]TronPayment, error) {
	var payments []TronPayment
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

// UpdateTronPayment updates an existing payment record
func (s *GormTronPaymentStore) UpdateTronPayment(payment *TronPayment) error {
	return s.db.Save(payment).Error
}

// GetPendingTronPayments retrieves all payments with "pending" status waiting for confirmation
func (s *GormTronPaymentStore) GetPendingTronPayments() ([]TronPayment, error) {
	var payments []TronPayment
	err := s.db.Where("status = ?", "pending").Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}
