// pkg/session/response_writer_test.go

package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponseWriter_WriteHeader(t *testing.T) {
	t.Run("should write header only once", func(t *testing.T) {
		rec := httptest.NewRecorder()
		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
			TTL:        1 * time.Hour,
		}

		sess, _ := NewSession(1 * time.Hour)

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
			sessionExists:  false,
			clearCookie:    false,
			statusWritten:  false,
		}

		rw.WriteHeader(http.StatusOK)
		assert.True(t, rw.statusWritten)
		assert.Equal(t, http.StatusOK, rec.Code)

		rw.WriteHeader(http.StatusInternalServerError)
		assert.Equal(t, http.StatusOK, rec.Code, "Status should not change after first WriteHeader")
	})

	t.Run("should set cookie when session exists", func(t *testing.T) {
		rec := httptest.NewRecorder()

		fixedTime := time.Date(2026, 1, 14, 15, 0, 0, 0, time.UTC)
		sess := &Session{
			SessionData: SessionData{
				ID:        "test-session-123",
				ExpiresAt: fixedTime.Add(1 * time.Hour),
			},
		}

		opts := &Options{
			CookieName: "session_id",
			Secret:     "secret-key",
			Path:       "/",
			TTL:        1 * time.Hour,
			HTTPOnly:   true,
			Secure:     true,
			SameSite:   http.SameSiteStrictMode,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
			sessionExists:  true,
			clearCookie:    false,
		}

		rw.WriteHeader(http.StatusOK)

		cookies := rec.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "session_id", cookie.Name)
		assert.NotEmpty(t, cookie.Value)
		assert.Equal(t, "/", cookie.Path)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		assert.Equal(t, 3600, cookie.MaxAge)
	})

	t.Run("should clear cookie when clearCookie is true", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
			HTTPOnly:   true,
			Secure:     false,
			SameSite:   http.SameSiteLaxMode,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			clearCookie:    true,
			sessionExists:  false,
		}

		rw.WriteHeader(http.StatusOK)

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "session_id", cookie.Name)
		assert.Empty(t, cookie.Value)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, -1, cookie.MaxAge)
		assert.True(t, cookie.Expires.Before(time.Now()))
	})

	t.Run("should not set cookie when session doesn't exist", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			sessionExists:  false,
			clearCookie:    false,
		}

		rw.WriteHeader(http.StatusOK)

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 0)
	})
}

func TestResponseWriter_Write(t *testing.T) {
	t.Run("should write response body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			sessionExists:  false,
		}

		body := []byte("Hello, World!")
		n, err := rw.Write(body)

		assert.NoError(t, err)
		assert.Equal(t, len(body), n)
		assert.Equal(t, "Hello, World!", rec.Body.String())
	})

	t.Run("should call WriteHeader(200) if not already called", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			sessionExists:  false,
			statusWritten:  false,
		}

		assert.False(t, rw.statusWritten)

		rw.Write([]byte("test"))

		assert.True(t, rw.statusWritten)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("should not call WriteHeader if already written", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			sessionExists:  false,
		}

		rw.WriteHeader(http.StatusCreated)
		assert.Equal(t, http.StatusCreated, rec.Code)

		rw.Write([]byte("test"))
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("should set cookie before writing body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		sess, _ := NewSession(1 * time.Hour)

		opts := &Options{
			CookieName: "session_id",
			Secret:     "secret",
			Path:       "/",
			TTL:        1 * time.Hour,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
			sessionExists:  true,
		}

		rw.Write([]byte("response body"))

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "session_id", cookies[0].Name)
	})
}

func TestResponseWriter_LoadCookieSession(t *testing.T) {
	t.Run("should set cookie with correct attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()

		fixedTime := time.Date(2026, 1, 14, 15, 0, 0, 0, time.UTC)

		sess := &Session{
			SessionData: SessionData{
				ID:        "session-abc123",
				ExpiresAt: fixedTime.Add(2 * time.Hour),
			},
		}

		opts := &Options{
			CookieName: "my_session",
			Secret:     "my-secret-key",
			Path:       "/app",
			TTL:        2 * time.Hour,
			HTTPOnly:   true,
			Secure:     true,
			SameSite:   http.SameSiteStrictMode,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
		}

		rw.loadCookieSession()

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "my_session", cookie.Name)
		assert.NotEmpty(t, cookie.Value)
		assert.Equal(t, "/app", cookie.Path)
		assert.Equal(t, fixedTime.Add(2*time.Hour), cookie.Expires)
		assert.Equal(t, 7200, cookie.MaxAge)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	})

	t.Run("should encode session ID with secret", func(t *testing.T) {
		rec := httptest.NewRecorder()

		sess := &Session{
			SessionData: SessionData{
				ID:        "test-id",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
		}

		opts := &Options{
			CookieName: "session_id",
			Secret:     "secret-key",
			Path:       "/",
			TTL:        1 * time.Hour,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
		}

		rw.loadCookieSession()

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)

		assert.Contains(t, cookies[0].Value, "s:")
		assert.Contains(t, cookies[0].Value, ".")
	})
}

func TestResponseWriter_ClearCookieSession(t *testing.T) {
	t.Run("should set cookie with MaxAge -1", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
			HTTPOnly:   true,
			Secure:     true,
			SameSite:   http.SameSiteStrictMode,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
		}

		rw.clearCookieSession()

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "session_id", cookie.Name)
		assert.Empty(t, cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	})

	t.Run("should set Expires to epoch", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
		}

		rw.clearCookieSession()

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)

		expectedExpires := time.Unix(0, 0).UTC()
		actualExpires := cookies[0].Expires.UTC()

		assert.Equal(t, expectedExpires, actualExpires)
	})

	t.Run("should respect cookie path", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/admin",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
		}

		rw.clearCookieSession()

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "/admin", cookies[0].Path)
	})
}

func TestResponseWriter_Integration(t *testing.T) {
	t.Run("should handle complete request lifecycle", func(t *testing.T) {
		rec := httptest.NewRecorder()

		sess, _ := NewSession(1 * time.Hour)
		sess.Set("user_id", "123")

		opts := &Options{
			CookieName: "session_id",
			Secret:     "secret",
			Path:       "/",
			TTL:        1 * time.Hour,
			HTTPOnly:   true,
			Secure:     false,
			SameSite:   http.SameSiteLaxMode,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
			sessionExists:  true,
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Success"))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Success", rec.Body.String())

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "session_id", cookies[0].Name)
		assert.NotEmpty(t, cookies[0].Value)
	})

	t.Run("should handle logout scenario", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			clearCookie:    true,
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Logged out"))

		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Empty(t, cookies[0].Value)
		assert.Equal(t, -1, cookies[0].MaxAge)
	})

	t.Run("should handle both set and clear (clear wins)", func(t *testing.T) {
		rec := httptest.NewRecorder()

		sess, _ := NewSession(1 * time.Hour)

		opts := &Options{
			CookieName: "session_id",
			Secret:     "secret",
			Path:       "/",
			TTL:        1 * time.Hour,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        sess,
			sessionExists:  true,
			clearCookie:    true,
		}

		rw.WriteHeader(http.StatusOK)

		cookies := rec.Result().Cookies()

		lastCookie := cookies[len(cookies)-1]
		assert.Empty(t, lastCookie.Value)
		assert.Equal(t, -1, lastCookie.MaxAge)
	})
}

func TestResponseWriter_EdgeCases(t *testing.T) {
	t.Run("should handle multiple Write calls", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
		}

		rw.Write([]byte("Hello "))
		rw.Write([]byte("World"))
		rw.Write([]byte("!"))

		assert.Equal(t, "Hello World!", rec.Body.String())
		assert.True(t, rw.statusWritten)
	})

	t.Run("should handle empty Write", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Path:       "/",
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
		}

		n, err := rw.Write([]byte{})

		assert.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.Empty(t, rec.Body.String())
	})

	t.Run("should handle nil session with sessionExists true", func(t *testing.T) {
		rec := httptest.NewRecorder()

		opts := &Options{
			CookieName: "session_id",
			Secret:     "secret",
			Path:       "/",
			TTL:        1 * time.Hour,
		}

		rw := &responseWriter{
			ResponseWriter: rec,
			opts:           opts,
			session:        nil, // ← nil session
			sessionExists:  true,
		}

		assert.Panics(t, func() {
			rw.WriteHeader(http.StatusOK)
		})
	})
}
