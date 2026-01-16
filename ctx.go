package session

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type contextKey string

const (
	sessionContextKey contextKey = "session"
)

type Context struct {
	*Session
	Token string
}

func newContext(sess *Session, secret string) *Context {
	ctx := &Context{
		Session: sess,
		Token:   encodeSessionId(sess.ID, secret),
	}

	sess.onHookRegenerate = func(newId string) {
		ctx.Token = encodeSessionId(newId, secret)
	}

	return ctx
}

func WithContext(ctx context.Context, sess *Context) context.Context {
	return context.WithValue(ctx, sessionContextKey, sess)
}

func MustFromContext(ctx context.Context) *Context {
	sess, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}

	return sess
}

func FromContext(ctx context.Context) (*Context, error) {
	val := ctx.Value(sessionContextKey)

	if val == nil {
		return nil, ErrSessionNotFound
	}

	if session, ok := val.(*Context); ok {
		return session, nil
	}
	return nil, ErrSessionInvalid
}

func encodeSessionId(id string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(id))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return "s:" + id + "." + sig
}
