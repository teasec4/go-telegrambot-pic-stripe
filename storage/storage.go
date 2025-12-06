package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type Storage interface {
	SavePhoto(userID string, fileID string) error
	GetPhotos(userID string) ([]string, error)
}

type JSONStorage struct {
	filepath string
	mu       sync.Mutex
}

type PhotoData struct {
	UserID  string   `json:"user_id"`
	FileIDs []string `json:"file_ids"`
}

func NewJSONStorage(filepath string) *JSONStorage {
	return &JSONStorage{
		filepath: filepath,
	}
}

func (s *JSONStorage) SavePhoto(userID string, fileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	photos, err := s.loadData()
	if err != nil {
		return err
	}

	// Ищем юзера
	found := false
	for i, p := range photos {
		if p.UserID == userID {
			photos[i].FileIDs = append(photos[i].FileIDs, fileID)
			found = true
			break
		}
	}

	// Если не нашли, создаём новую запись
	if !found {
		photos = append(photos, PhotoData{
			UserID:  userID,
			FileIDs: []string{fileID},
		})
	}

	return s.saveData(photos)
}

func (s *JSONStorage) GetPhotos(userID string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	photos, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, p := range photos {
		if p.UserID == userID {
			return p.FileIDs, nil
		}
	}

	return []string{}, nil
}

func (s *JSONStorage) loadData() ([]PhotoData, error) {
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		return []PhotoData{}, nil
	}

	data, err := ioutil.ReadFile(s.filepath)
	if err != nil {
		return nil, err
	}

	var photos []PhotoData
	err = json.Unmarshal(data, &photos)
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func (s *JSONStorage) saveData(photos []PhotoData) error {
	data, err := json.MarshalIndent(photos, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.filepath, data, 0644)
}
