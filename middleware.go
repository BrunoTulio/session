package session

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

type Middleware struct {
	log               Logger
	store             Store
	cookieName        string
	secret            string
	ttl               time.Duration
	httpOnly          bool
	secure            bool
	sameSite          http.SameSite
	saveUninitialized bool
	autoRenew         bool
	path              string
	//handleError       ErrorHandler
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.loadSession(r)
		if err != nil {
			m.log.Warnf("session id resolve failed: %v", err)
		}
		if session == nil && m.saveUninitialized {
			session = NewSession(m.ttl)
			m.log.Debugf("Anonymous session created: sessionID=%s",
				session.ID[:8]+"...",
			)
		}

		ctx := WithContext(r.Context(), session)
		ww := m.writerWithCookie(w, session, err)
		next.ServeHTTP(ww, r.WithContext(ctx))
		m.storeSession(ctx, session)
	})
}

func Handler(opts ...func(*Options)) func(handler http.Handler) http.Handler {
	opt := &Options{
		Logger:            &defaultLogger{},
		Store:             NewMemoryStore(),
		SaveUninitialized: false,
		AutoRenew:         false,
		Secret:            "secret",
		CookieName:        "sid",
		Path:              "/",
		HTTPOnly:          true,
		Secure:            false,
		SameSite:          http.SameSiteNoneMode,
		TTL:               time.Hour * 1,
		//HandleError:       nil,
	}

	for _, o := range opts {
		o(opt)
	}

	m := &Middleware{
		log:               opt.Logger,
		store:             opt.Store,
		cookieName:        opt.CookieName,
		secret:            opt.Secret,
		ttl:               opt.TTL,
		httpOnly:          opt.HTTPOnly,
		secure:            opt.Secure,
		sameSite:          opt.SameSite,
		saveUninitialized: opt.SaveUninitialized,
		autoRenew:         opt.AutoRenew,
		path:              opt.Path,
		//handleError:       opt.HandleError,
	}

	return m.Handler
}

func HandlerWithOptions(opt Options) func(http.Handler) http.Handler {
	return Handler(
		WithLogger(opt.Logger),
		WithStore(opt.Store),
		WithCookieName(opt.CookieName),
		WithSecret(opt.Secret),
		WithTTL(opt.TTL),
		WithHTTPOnly(opt.HTTPOnly),
		WithSecure(opt.Secure),
		WithSameSite(opt.SameSite),
		WithSaveUninitialized(opt.SaveUninitialized),
		WithAutoRenew(opt.AutoRenew),
		WithPath(opt.Path),
		//WithHandleError(opt.HandleError),
	)
}

func (m *Middleware) storeSession(ctx context.Context, session *Session) {
	shouldSave := session != nil && session.IsModified()

	if !shouldSave {
		return
	}
	m.cleanupOldSession(session)
	if err := m.store.Set(ctx, session.SessionData); err != nil {
		m.log.Errorf("Failed to set session: %v", err)
	}
}

func (m *Middleware) cleanupOldSession(session *Session) {
	if !session.HasOldID() {
		return
	}

	oldID := session.oldID
	go func() {
		ctx := context.Background()
		if err := m.store.Delete(ctx, oldID); err != nil {
			m.log.Warnf("Failed to delete old session %s: %v", oldID[:8]+"...", err)
		} else {
			m.log.Debugf("Old session deleted: %s", oldID[:8]+"...")
		}
	}()
	session.clearOldID()
}

func (m *Middleware) writerWithCookie(w http.ResponseWriter, session *Session, err error) *responseWriter {
	ww := &responseWriter{
		ResponseWriter: w,
		statusWritten:  false,
		before: func() {
			if err != nil && !errors.Is(err, ErrNoCookie) {
				http.SetCookie(w, &http.Cookie{
					Name:     m.cookieName,
					Value:    "",
					Path:     m.path,
					Secure:   m.secure,
					HttpOnly: m.httpOnly,
					SameSite: m.sameSite,
					MaxAge:   -1,
					Expires:  time.Unix(0, 0),
				})
				return
			}

			if session != nil {
				http.SetCookie(w, &http.Cookie{
					Name:     m.cookieName,
					Value:    session.SignedID(m.secret),
					Path:     m.path,
					Expires:  session.ExpiresAt,
					MaxAge:   int(m.ttl.Seconds()),
					Secure:   m.secure,
					HttpOnly: m.httpOnly,
					SameSite: m.sameSite,
				})
			}
		},
	}

	return ww
}

//func (m *Middleware) writeError(err error, w http.ResponseWriter, r *http.Request) {
//	m.log.Errorf("Failed to create session: %v", err)
//	if m.handleError != nil {
//		m.handleError(w, r, http.StatusInternalServerError, err)
//		return
//	}
//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//}

func (m *Middleware) loadSession(r *http.Request) (*Session, error) {
	ctx := r.Context()
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, ErrNoCookie
		}
		m.log.Debugf("Cookie read error: cookieName=%s, error=%v, path=%s",
			m.cookieName,
			err,
			r.URL.Path,
		)
		return nil, ErrInvalidCookie
	}

	if cookie.Value == "" {
		return nil, ErrInvalidCookie
	}

	sessionID, err := m.unsignCookie(cookie.Value)
	if err != nil {
		return nil, err
	}

	data, err := m.store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	session := NewSessionFromData(data)

	if session.IsExpired() {
		m.log.Warnf("session [%s] expired", session.ID)
		go func() {
			if err := m.store.Delete(context.Background(), sessionID); err != nil {
				m.log.Warnf("session delete session failed: %v", err)
			}
		}()
		return nil, ErrSessionExpired
	}

	if m.autoRenew {
		session.Renew(m.ttl)
	}

	return session, nil
}

func (m *Middleware) unsignCookie(signedValue string) (string, error) {
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

	h := hmac.New(sha256.New, []byte(m.secret))
	h.Write([]byte(sessionID))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(receivedSig), []byte(expectedSig)) {
		return "", ErrInvalidSignature
	}

	return sessionID, nil
}
