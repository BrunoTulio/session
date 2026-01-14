package session

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Middleware(opts ...func(*Options)) func(handler http.Handler) http.Handler {
	opt := &Options{
		Logger:            &defaultLogger{},
		Store:             NewMemoryStore(),
		CookieName:        "sid",
		Secret:            "secret",
		TTL:               time.Hour * 1,
		HTTPOnly:          true,
		Secure:            false,
		SameSite:          http.SameSiteNoneMode,
		SaveUninitialized: false,
		AutoRenew:         false,
	}

	for _, o := range opts {
		o(opt)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clearCookie, session, err := loadSession(r, opt)
			if err != nil {
				opt.Logger.Warnf("session id resolve failed: %v", err)
			}

			if session == nil && opt.SaveUninitialized {
				session, err = NewSession(opt.TTL)
				if err != nil {
					opt.Logger.Errorf("Failed to create session: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				opt.Logger.Debugf("Anonymous session created: sessionID=%s",
					session.ID[:8]+"...",
				)
			}

			exist := session != nil
			ctx := WithContext(r.Context(), session)

			ww := &responseWriter{
				ResponseWriter: w,
				opts:           opt,
				session:        session,
				sessionExists:  exist,
				clearCookie:    clearCookie,
			}

			next.ServeHTTP(ww, r.WithContext(ctx))
			shouldSave := exist && session.IsModified()
			if shouldSave {
				if err := opt.Store.Set(ctx, session.SessionData); err != nil {
					opt.Logger.Errorf("Failed to set session: %v", err)
				}
			}
		})
	}
}

func loadSession(r *http.Request, opt *Options) (bool, *Session, error) {
	ctx := r.Context()
	store := opt.Store
	secret := opt.Secret
	log := opt.Logger
	ttl := opt.TTL

	fmt.Println("Cookie header:", r.Header.Get("Cookie"))
	fmt.Println("All headers:", r.Header)

	cookie, err := r.Cookie(opt.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return false, nil, nil
		}

		log.Debugf("Cookie read error: cookieName=%s, error=%v, path=%s",
			opt.CookieName,
			err,
			r.URL.Path,
		)
		return false, nil, ErrInvalidCookie
	}

	if cookie.Value == "" {
		return true, nil, ErrInvalidCookie
	}

	sessionID, err := unsignCookie(cookie.Value, secret)
	if err != nil {
		return true, nil, err
	}

	data, err := store.Get(ctx, sessionID)
	if err != nil {
		return false, nil, err
	}

	session := NewSessionFromData(data)

	if session.IsExpired() {
		log.Warnf("session [%s] expired", session.ID)
		go func() {
			if err := store.Delete(context.Background(), sessionID); err != nil {
				log.Warnf("session delete session failed: %v", err)
			}
		}()
		return true, nil, ErrSessionExpired
	}

	if opt.AutoRenew {
		session.Renew(ttl)
	}

	return false, session, nil
}

func unsignCookie(signedValue string, secret string) (string, error) {
	if !strings.HasPrefix(signedValue, "s:") || len(signedValue) < 2 {
		return "", ErrInvalidSignature
	}

	value := signedValue[2:]

	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", ErrInvalidSignature
	}

	sessionID := parts[0]
	receivedSig := parts[1]

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(sessionID))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(receivedSig), []byte(expectedSig)) {
		return "", ErrInvalidSignature
	}

	return sessionID, nil
}
