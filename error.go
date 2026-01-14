package session

import "errors"

var (
	ErrInvalidCookie       = errors.New("invalid session cookie")
	ErrInvalidSignature    = errors.New("invalid cookie signature")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionInvalid      = errors.New("invalid session")
	ErrSessionIDGeneration = errors.New("failed to generate session id")
)
