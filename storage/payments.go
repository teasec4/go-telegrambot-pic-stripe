package storage

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"` // "paid", "image_sent", "failed"
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaymentStore interface {
	SavePayment(payment *Payment) error
	GetPayment(id string) (*Payment, error)
	UpdatePaymentStatus(id string, status string) error
	GetFailedPayments() ([]Payment, error)
}

type GormPaymentStore struct {
	db *gorm.DB
}

func NewGormPaymentStore(db *gorm.DB) *GormPaymentStore {
	db.AutoMigrate(&Payment{})
	return &GormPaymentStore{db: db}
}

func (s *GormPaymentStore) SavePayment(payment *Payment) error {
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	return s.db.Create(payment).Error
}

func (s *GormPaymentStore) GetPayment(id string) (*Payment, error) {
	var payment Payment
	err := s.db.First(&payment, "id = ?", id).Error
	return &payment, err
}

func (s *GormPaymentStore) UpdatePaymentStatus(id string, status string) error {
	return s.db.Model(&Payment{}).Where("id = ?", id).Update("status", status).Error
}

func (s *GormPaymentStore) GetFailedPayments() ([]Payment, error) {
	var payments []Payment
	err := s.db.Where("status = ?", "failed").Find(&payments).Error
	return payments, err
}
