package session

import "errors"

var (
	ErrNoCookie            = errors.New("no cookie found")
	ErrInvalidCookie       = errors.New("invalid session cookie")
	ErrInvalidSignature    = errors.New("invalid cookie signature")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionInvalid      = errors.New("invalid session")
	ErrSessionIDGeneration = errors.New("failed to generate session id")
)
