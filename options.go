package session

import (
	"net/http"
	"time"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request, status int, err error)

type Options struct {
	Logger            Logger
	Store             Store
	SaveUninitialized bool
	AutoRenew         bool
	Secret            string
	CookieName        string
	Path              string
	HTTPOnly          bool
	Secure            bool
	SameSite          http.SameSite
	TTL               time.Duration
}

func WithLogger(logger Logger) func(*Options) {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithStore(store Store) func(*Options) {
	return func(o *Options) {
		o.Store = store
	}
}

func WithCookieName(name string) func(*Options) {
	return func(o *Options) {
		o.CookieName = name
	}
}

func WithSecret(secret string) func(*Options) {
	return func(o *Options) {
		o.Secret = secret
	}
}

func WithTTL(ttl time.Duration) func(*Options) {
	return func(o *Options) {
		o.TTL = ttl
	}
}

func WithHTTPOnly(httpOnly bool) func(*Options) {
	return func(o *Options) {
		o.HTTPOnly = httpOnly
	}
}

func WithSecure(secure bool) func(*Options) {
	return func(o *Options) {
		o.Secure = secure
	}
}

func WithSameSite(sameSite http.SameSite) func(*Options) {
	return func(o *Options) {
		o.SameSite = sameSite
	}
}

func WithSaveUninitialized(saveUninitialized bool) func(*Options) {
	return func(o *Options) {
		o.SaveUninitialized = saveUninitialized
	}
}

func WithAutoRenew(autoRenew bool) func(*Options) {
	return func(o *Options) {
		o.AutoRenew = autoRenew
	}
}

func WithPath(path string) func(*Options) {
	return func(o *Options) {
		o.Path = path
	}
}
