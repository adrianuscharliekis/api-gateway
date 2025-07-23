package utils

import (
	"sync"
	"time"
)

type ExternalIDStore struct {
	data sync.Map
}

func NewExternalIDStore() *ExternalIDStore {
	return &ExternalIDStore{}
}

// Save stores the externalID with expiry set to midnight of the next day
func (s *ExternalIDStore) Save(id string) {
	// Midnight of the next day
	midnight := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	s.data.Store(id, midnight)
}

// ExistsAndValid checks if the externalID exists and is still valid
func (s *ExternalIDStore) ExistsAndValid(id string) bool {
	value, ok := s.data.Load(id)
	if !ok {
		return false
	}
	expiry := value.(time.Time)
	if time.Now().After(expiry) {
		s.data.Delete(id)
		return false
	}
	return true
}
