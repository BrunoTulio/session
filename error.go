package session

import "errors"

var (
	ErrNoCookie         = errors.New("no cookie found")
	ErrInvalidCookie    = errors.New("invalid session cookie")
	ErrInvalidSignature = errors.New("invalid cookie signature")
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionExpired   = errors.New("session expired")

	ErrStoreNotFound = errors.New("store not found in context")
	ErrStoreInvalid  = errors.New("invalid store in context")
)
