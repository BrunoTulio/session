package session

import "context"

type contextKey string

const (
	sessionContextKey contextKey = "session"
)

type Context struct {
	*Session
	Token string
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
