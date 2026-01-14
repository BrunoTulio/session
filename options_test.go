package session_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/BrunoTulio/session"
	"github.com/BrunoTulio/session/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	t.Run("should set logger", func(t *testing.T) {
		opts := &session.Options{}
		logger := mocks.NewMockLogger()

		fn := session.WithLogger(logger)
		fn(opts)

		assert.NotNil(t, opts.Logger)
		assert.Equal(t, logger, opts.Logger)
	})
}

func TestWithStore(t *testing.T) {
	t.Run("should set store", func(t *testing.T) {
		opts := &session.Options{}
		store := mocks.NewMockStore()

		fn := session.WithStore(store)
		fn(opts)

		assert.NotNil(t, opts.Store)
		assert.Equal(t, store, opts.Store)
	})
}

func TestWithCookieName(t *testing.T) {
	t.Run("should set cookie_name", func(t *testing.T) {
		opts := &session.Options{}
		cookiesNames := []string{"cookie_test_1", "cookie_test_2"}
		for _, cookieName := range cookiesNames {

			fn := session.WithCookieName(cookieName)
			fn(opts)
			assert.Equal(t, cookieName, opts.CookieName)
		}
	})
}

func TestWithSecret(t *testing.T) {
	t.Run("should set secret", func(t *testing.T) {
		opts := &session.Options{}

		secrets := []string{"secret_1", "secret_2"}

		for _, secret := range secrets {
			fn := session.WithSecret(secret)
			fn(opts)
			assert.Equal(t, secret, opts.Secret)
		}
	})
}

func TestWithTTL(t *testing.T) {
	t.Run("should set ttl", func(t *testing.T) {
		opts := &session.Options{}

		ttls := []time.Duration{time.Second * 1, time.Second * 2}

		for _, ttl := range ttls {
			fn := session.WithTTL(ttl)
			fn(opts)
			assert.Equal(t, ttl, opts.TTL)
		}
	})
}

func TestWithHTTPOnly(t *testing.T) {
	t.Run("should set http only", func(t *testing.T) {
		opts := &session.Options{}

		values := []bool{true, false}

		for _, val := range values {
			fn := session.WithHTTPOnly(val)
			fn(opts)
			assert.Equal(t, val, opts.HTTPOnly)
		}
	})
}

func TestWithSecure(t *testing.T) {
	t.Run("should set secure", func(t *testing.T) {
		opts := &session.Options{}

		values := []bool{true, false}

		for _, val := range values {
			fn := session.WithSecure(val)
			fn(opts)
			assert.Equal(t, val, opts.Secure)
		}
	})
}

func TestWithSameSite(t *testing.T) {
	t.Run("should set same site", func(t *testing.T) {
		opts := &session.Options{}

		values := []http.SameSite{
			http.SameSiteDefaultMode,
			http.SameSiteLaxMode,
			http.SameSiteStrictMode,
			http.SameSiteNoneMode,
		}

		for _, val := range values {
			fn := session.WithSameSite(val)
			fn(opts)
			assert.Equal(t, val, opts.SameSite)
		}
	})
}

func TestWithSaveUninitialized(t *testing.T) {
	t.Run("should set save uninitialized", func(t *testing.T) {
		opts := &session.Options{}

		values := []bool{true, false}

		for _, val := range values {
			fn := session.WithSaveUninitialized(val)
			fn(opts)
			assert.Equal(t, val, opts.SaveUninitialized)
		}
	})
}

func TestWithAutoRenew(t *testing.T) {
	t.Run("should set autorenew", func(t *testing.T) {
		autoRenews := []bool{true, false}

		opt := &session.Options{AutoRenew: false}

		for _, autoRenew := range autoRenews {
			fn := session.WithAutoRenew(autoRenew)
			fn(opt)
			assert.Equal(t, autoRenew, opt.AutoRenew)
		}
	})
}

func TestWithPath(t *testing.T) {
	opts := &session.Options{}

	values := []string{"path_1", "path_2", "path_3"}

	for _, val := range values {
		fn := session.WithPath(val)
		fn(opts)
		assert.Equal(t, val, opts.Path)
	}
}
