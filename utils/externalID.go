package utils

import (
	"sync"
	"time"
)

type ExternalIDStore struct {
	data sync.Map
	ttl  time.Duration
}

func NewExternalIDStore(ttl time.Duration) *ExternalIDStore {
	return &ExternalIDStore{ttl: ttl}
}

// Save stores the externalID with a TTL
func (s *ExternalIDStore) Save(id string) {
	s.data.Store(id, time.Now().Add(s.ttl))
}

// ExistsAndValid checks if the externalID exists and is still valid
func (s *ExternalIDStore) ExistsAndValid(id string) bool {
	value, ok := s.data.Load(id)
	if !ok {
		return false
	}
	expiry := value.(time.Time)
	if time.Now().After(expiry) {
		s.data.Delete(id) // clean up expired
		return false
	}
	return true
}
