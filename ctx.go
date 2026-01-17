package session

import (
	"context"
	"time"
)

type contextKey string

const (
	sessionContextKey contextKey = "session"
)

func GetOrCreate(ctx context.Context, ttl time.Duration) *Session {
	holder := getHolderContext(ctx)
	if holder == nil {
		panic("session: middleware not configured")
	}

	sess := holder.get()
	if sess != nil {
		return sess
	}
	sess = NewSession(ttl)
	holder.set(sess)
	return sess
}

func HasSession(ctx context.Context) bool {
	_, err := FromContext(ctx)
	return err == nil
}

func MustFromContext(ctx context.Context) *Session {
	sess, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}

	return sess
}

func FromContext(ctx context.Context) (*Session, error) {
	holder := getHolderContext(ctx)
	if holder == nil {
		return nil, ErrSessionNotFound
	}

	sess := holder.get()
	if sess == nil {
		return nil, ErrSessionNotFound
	}

	return sess, nil
}

func withHolderContext(ctx context.Context, holder *holder) context.Context {
	return context.WithValue(ctx, sessionContextKey, holder)
}

func getHolderContext(ctx context.Context) *holder {
	val := ctx.Value(sessionContextKey)
	if val == nil {
		return nil
	}
	if holder, ok := val.(*holder); ok {
		return holder
	}
	return nil
}
