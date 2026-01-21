package session

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"sync"
	"time"
)

const (
	sessionIDBytes = 24
)

type Session struct {
	SessionData
	modified  bool
	oldID     string
	destroyed bool
	isNew     bool
	mu        sync.RWMutex
}

func NewSession(ttl time.Duration) *Session {
	data := NewSessionData(ttl)

	return &Session{
		SessionData: data,
		modified:    true,
		isNew:       true,
	}
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

func (s *Session) Set(key string, value any) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Set(key, value)
	return s
}

func (s *Session) Delete(key string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Delete(key)
	return s
}

func (s *Session) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

func (s *Session) MarkClean() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.modified = false
	return s
}

func (s *Session) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionData.IsExpired()
}

func (s *Session) Renew(ttl time.Duration) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Renew(ttl)

	return s
}

func (s *Session) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Authenticated
}

func (s *Session) Authenticate(userID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Authenticate(userID)

	return s
}

func (s *Session) Unauthenticate() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modified = true
	s.SessionData.Unauthenticate()
	return s
}

func (s *Session) HasOldID() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oldID != ""
}

func (s *Session) Regenerate() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	newID := generateId()
	s.oldID = s.ID
	s.ID = newID
	s.modified = true
	return s
}
func (s *Session) SignedID(secret string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return encodeSessionId(s.ID, secret)
}

func (s *Session) GetOldID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oldID
}

func (s *Session) Destroy() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.destroyed = true
	s.modified = true
}

func (s *Session) IsNew() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isNew
}

func (s *Session) IsDestroyed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.destroyed
}

func (s *Session) Flush(ctx context.Context) error {
	store, err := GetStore(ctx)

	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.modified {
		return nil
	}

	if err := store.Set(ctx, s.SessionData); err != nil {
		return err
	}

	s.modified = false
	return nil
}

func (s *Session) markPersisted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isNew = false
}

func (s *Session) clearOldID() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.oldID = ""
	return s
}

func encodeSessionId(id string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(id))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return "s:" + id + "." + sig
}
