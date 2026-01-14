package session

import (
	"context"
)

type Store interface {
	Get(ctx context.Context, id string) (SessionData, error)
	Set(ctx context.Context, session SessionData) error
	Delete(ctx context.Context, id string) error
}
