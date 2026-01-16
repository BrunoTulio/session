package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithContext_(t *testing.T) {
	t.Run("should save the session", func(t *testing.T) {
		now := time.Now()
		s := Session{
			SessionData: SessionData{
				ID:        "id",
				Data:      make(map[string]any),
				CreatedAt: now,
				ExpiresAt: now,
				UpdatedAt: now,
			},
		}

		sessCtx := Context{
			Session: &s,
			Token:   "",
		}

		ctx := context.Background()
		ctx = WithContext(ctx, &sessCtx)

		retrieved, err := FromContext(ctx)

		assert.NoError(t, err)
		assert.Equal(t, &sessCtx, retrieved)
		assert.Equal(t, now, retrieved.CreatedAt)
		assert.Equal(t, now, retrieved.ExpiresAt)
		assert.Equal(t, now, retrieved.UpdatedAt)
		assert.Equal(t, s.ID, retrieved.ID)
	})
	t.Run("should save the session nil", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithContext(ctx, nil)

		s, err := FromContext(ctx)

		assert.Nil(t, s)
		assert.NoError(t, err)
	})
}

func TestFromContext(t *testing.T) {
	t.Run("should return the session successfully", func(t *testing.T) {
		now := time.Now()

		s := Session{
			SessionData: SessionData{
				ID:        "id",
				Data:      make(map[string]any),
				CreatedAt: now,
				ExpiresAt: now,
				UpdatedAt: now,
			},
		}

		sessCtx := Context{
			Session: &s,
			Token:   "token",
		}

		ctx := context.Background()
		ctx = WithContext(ctx, &sessCtx)

		retrieved, err := FromContext(ctx)

		assert.NoError(t, err)
		assert.Equal(t, &sessCtx, retrieved)
		assert.Equal(t, now, retrieved.CreatedAt)
		assert.Equal(t, now, retrieved.ExpiresAt)
		assert.Equal(t, now, retrieved.UpdatedAt)
		assert.Equal(t, s.ID, retrieved.ID)
		assert.Equal(t, "token", sessCtx.Token)
	})
	t.Run("should return error session not fount", func(t *testing.T) {
		ctx := context.Background()

		retrieved, err := FromContext(ctx)

		assert.Error(t, err)
		assert.Equal(t, err, ErrSessionNotFound)
		assert.Nil(t, retrieved)
	})
	t.Run("should return error session invalid", func(t *testing.T) {
		ctx := context.Background()
		ctx = testSetContextValue(ctx, "invalid")

		retrieved, err := FromContext(ctx)

		assert.Error(t, err)
		assert.Equal(t, err, ErrSessionInvalid)
		assert.Nil(t, retrieved)
	})
}

func TestMustFromContext(t *testing.T) {
	t.Run("should panic if session context value", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = MustFromContext(context.Background())
		})
	})

	t.Run("should return the session success", func(t *testing.T) {
		now := time.Now()

		s := Session{
			SessionData: SessionData{
				ID:        "id",
				Data:      make(map[string]any),
				CreatedAt: now,
				ExpiresAt: now,
				UpdatedAt: now,
			},
		}

		sessCtx := Context{
			Session: &s,
			Token:   "token",
		}

		ctx := context.Background()
		ctx = WithContext(ctx, &sessCtx)

		retrieved := MustFromContext(ctx)

		assert.Equal(t, &sessCtx, retrieved)
		assert.Equal(t, now, retrieved.CreatedAt)
		assert.Equal(t, now, retrieved.ExpiresAt)
		assert.Equal(t, now, retrieved.UpdatedAt)
		assert.Equal(t, s.ID, retrieved.ID)
	})

	t.Run("should panic if session not fount", func(t *testing.T) {
		assert.Panics(t, func() {
			ctx := context.Background()

			retrieved := MustFromContext(ctx)

			assert.Nil(t, retrieved)
		})
	})

	t.Run("should panic if session invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			ctx := context.Background()
			ctx = testSetContextValue(ctx, "invalid")

			retrieved := MustFromContext(ctx)

			assert.Nil(t, retrieved)
		})
	})
}
