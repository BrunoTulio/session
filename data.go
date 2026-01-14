package session

import "time"

type SessionData struct {
	ID            string         `json:"id"`
	Data          map[string]any `json:"data"`
	CreatedAt     time.Time      `json:"created_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Authenticated bool           `json:"authenticated"`
	UserID        string         `json:"user_id"`
}

func NewSessionData(ttl time.Duration) (SessionData, error) {
	id, err := generateId()
	if err != nil {
		return SessionData{}, err
	}
	t := now()
	return SessionData{
		ID:        id,
		CreatedAt: t,
		UpdatedAt: t,
		ExpiresAt: t.Add(ttl),
		Data:      make(map[string]any),
	}, nil
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
