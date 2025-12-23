package storer

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type GormStorer struct {
	db *gorm.DB
}

func NewGormStorer(db *gorm.DB) *GormStorer {
	db.AutoMigrate(&Payment{}, &Photo{})
	return &GormStorer{db: db}
}

// ========== Payments - Generic ==========

func (s *GormStorer) savePaymentWithType(payment *Payment, paymentType string) error {
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	payment.Type = paymentType
	return s.db.Create(payment).Error
}

func (s *GormStorer) getPaymentByField(field, value, paymentType string) (*Payment, error) {
	var payment Payment
	err := s.db.Where("? = ? AND type = ?", gorm.Expr(field), value, paymentType).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *GormStorer) getPaymentsByStatus(status, paymentType string) ([]Payment, error) {
	var payments []Payment
	err := s.db.Where("status = ? AND type = ?", status, paymentType).Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

// ========== Payments - Stripe ==========

func (s *GormStorer) SavePayment(payment *Payment) error {
	return s.savePaymentWithType(payment, "stripe")
}

func (s *GormStorer) GetPayment(id string) (*Payment, error) {
	return s.getPaymentByField("id", id, "stripe")
}

func (s *GormStorer) UpdatePaymentStatus(id string, status string) error {
	return s.db.Model(&Payment{}).Where("id = ?", id).Update("status", status).Error
}

func (s *GormStorer) GetFailedPayments() ([]Payment, error) {
	return s.getPaymentsByStatus("failed", "stripe")
}

// ========== Payments - Tron ==========

func (s *GormStorer) SaveTronPayment(payment *Payment) error {
	return s.savePaymentWithType(payment, "tron")
}

func (s *GormStorer) GetTronPayment(txID string) (*Payment, error) {
	return s.getPaymentByField("tx_id", txID, "tron")
}

func (s *GormStorer) GetTronPaymentByAddress(address string) (*Payment, error) {
	return s.getPaymentByField("address", address, "tron")
}

func (s *GormStorer) GetTronPaymentsByUserID(userID string) ([]Payment, error) {
	var payments []Payment
	err := s.db.Where("user_id = ? AND type = ?", userID, "tron").Order("created_at DESC").Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s *GormStorer) UpdateTronPayment(payment *Payment) error {
	return s.db.Save(payment).Error
}

func (s *GormStorer) GetPendingTronPayments() ([]Payment, error) {
	return s.getPaymentsByStatus("pending", "tron")
}

// ========== Photos ==========

func (s *GormStorer) SavePhoto(photo *Photo) error {
	photo.CreatedAt = time.Now()
	return s.db.Create(photo).Error
}

func (s *GormStorer) GetRandomPhoto() (*Photo, error) {
	var photos []Photo
	err := s.db.Find(&photos).Error
	if err != nil {
		return nil, err
	}
	if len(photos) == 0 {
		return nil, nil
	}

	randomIndex := rand.Intn(len(photos))
	return &photos[randomIndex], nil
}
