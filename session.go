package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	sessionIDBytes = 24
)

type Session struct {
	SessionData
	modified bool
	oldID    string
	mu       sync.RWMutex
}

func NewSession(ttl time.Duration) (*Session, error) {
	data, err := NewSessionData(ttl)
	if err != nil {
		return nil, err
	}

	return &Session{
		SessionData: data,
		modified:    true,
	}, nil
}

func (s *Session) GetSessionData() SessionData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionData
}

func NewSessionFromData(sd SessionData) *Session {
	if sd.Data == nil {
		sd.Data = make(map[string]any)
	}

	return &Session{
		SessionData: sd,
	}
}

func (s *Session) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionData.Get(key)
}

func (s *Session) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Set(key, value)
}

func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Delete(key)
}

func (s *Session) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

func (s *Session) MarkClean() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.modified = false
}

func (s *Session) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionData.IsExpired()
}

func (s *Session) Renew(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Renew(ttl)
}

func (s *Session) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Authenticated
}

func (s *Session) Authenticate(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Authenticate(userID)
}

func (s *Session) Unauthenticate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Unauthenticate()
}

func (s *Session) encodeSessionId(secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(s.ID))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return "s:" + s.ID + "." + sig
}

func (s *Session) HasOldSessionID() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oldID != ""
}

func (s *Session) Regenerate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newID, err := generateId()
	if err != nil {
		return fmt.Errorf("failed to generate new session ID: %w", err)
	}
	s.oldID = s.ID
	s.ID = newID
	s.modified = true

	return nil
}

func (s *Session) ClearOldSessionID() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.oldID = ""
}

func generateId() (string, error) {
	bytes := make([]byte, sessionIDBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("%w: %v", ErrSessionIDGeneration, err)
	}

	return hex.EncodeToString(bytes), nil
}
