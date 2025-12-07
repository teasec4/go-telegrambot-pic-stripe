package storage

import (
	"time"
	"math/rand"
	"gorm.io/gorm"
)

type PhotoStore interface {
	SavePhoto(photo *Photo) error
	GetRandomPhoto() (*Photo, error)
}

type Photo struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	FileID   string `json:"file_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GromPhotoStore struct{
	db *gorm.DB
}

func NewGormPhotoStore(db *gorm.DB) *GromPhotoStore{
	db.AutoMigrate(&Photo{})
	return &GromPhotoStore{db: db}
}

func (s *GromPhotoStore) SavePhoto(photo *Photo) error{
	photo.CreatedAt = time.Now()
	return s.db.Create(photo).Error
}


func (s *GromPhotoStore) GetRandomPhoto() (*Photo, error) {
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


