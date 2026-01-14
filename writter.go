package session

import (
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	opts          *Options
	session       *Session
	sessionExists bool
	clearCookie   bool
	statusWritten bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusWritten {
		return
	}

	if rw.sessionExists {
		rw.loadCookieSession()
	}

	if rw.clearCookie {
		rw.clearCookieSession()
	}

	rw.statusWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) clearCookieSession() {
	http.SetCookie(rw, &http.Cookie{
		Name:     rw.opts.CookieName,
		Value:    "",
		Path:     rw.opts.Path,
		Secure:   rw.opts.Secure,
		HttpOnly: rw.opts.HTTPOnly,
		SameSite: rw.opts.SameSite,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (rw *responseWriter) loadCookieSession() {
	http.SetCookie(rw, &http.Cookie{
		Name:     rw.opts.CookieName,
		Value:    rw.session.encodeSessionId(rw.opts.Secret),
		Path:     rw.opts.Path,
		Expires:  rw.session.ExpiresAt,
		MaxAge:   int(rw.opts.TTL.Seconds()),
		Secure:   rw.opts.Secure,
		HttpOnly: rw.opts.HTTPOnly,
		SameSite: rw.opts.SameSite,
	})
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.statusWritten {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
