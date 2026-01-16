package session

import "context"

type contextKey string

const (
	sessionContextKey contextKey = "session"
)

func WithContext(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, sess)
}

func MustFromContext(ctx context.Context) *Session {
	sess, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}

	return sess
}

func FromContext(ctx context.Context) (*Session, error) {
	val := ctx.Value(sessionContextKey)

	if val == nil {
		return nil, ErrSessionNotFound
	}

	if session, ok := val.(*Session); ok {
		return session, nil
	}
	return nil, ErrSessionInvalid
}
