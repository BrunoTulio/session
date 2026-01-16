package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type SessionData struct {
	ID            string         `json:"id"`
	Data          map[string]any `json:"data"`
	CreatedAt     time.Time      `json:"created_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Authenticated bool           `json:"authenticated"`
	UserID        string         `json:"user_id"`
}

func NewSessionData(ttl time.Duration) SessionData {
	id := generateId()

	t := now()
	return SessionData{
		ID:        id,
		CreatedAt: t,
		UpdatedAt: t,
		ExpiresAt: t.Add(ttl),
		Data:      make(map[string]any),
	}
}

func (s *SessionData) Get(key string) (any, bool) {
	val, ok := s.Data[key]
	return val, ok
}

func (s *SessionData) Set(key string, value any) {
	s.Data[key] = value
	s.UpdatedAt = now()
}

func (s *SessionData) Delete(key string) {
	delete(s.Data, key)
	s.UpdatedAt = now()
}

func (s *SessionData) Authenticate(userID string) {
	s.UpdatedAt = now()
	s.Authenticated = true
	s.UserID = userID
}

func (s *SessionData) Unauthenticate() {
	s.UpdatedAt = now()
	s.Authenticated = false
	s.UserID = ""
}

func (s *SessionData) Renew(ttl time.Duration) {
	now := now()
	s.ExpiresAt = now.Add(ttl)
	s.UpdatedAt = now
}

func (s *SessionData) IsExpired() bool {
	return now().After(s.ExpiresAt)
}

// generateID generates a cryptographically secure session ID.
//
// In Go 1.24+, crypto/rand.Read never returns an error. If random number
// generation fails, the program crashes via fatal() as it's unsafe to
// continue without secure randomness. See: https://go.dev/issue/66821
func generateId() string {
	bytes := make([]byte, sessionIDBytes)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
