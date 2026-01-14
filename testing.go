package session

import "context"

func testSetContextValue(ctx context.Context, value any) context.Context {
	return context.WithValue(ctx, sessionContextKey, value)
}
