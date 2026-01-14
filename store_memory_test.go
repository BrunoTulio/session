package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStore(t *testing.T) {
	t.Run("should create new memory store", func(t *testing.T) {
		store := NewMemoryStore()

		assert.NotNil(t, store)
	})
}

func TestMemoryStore_Set(t *testing.T) {
	t.Run("should store session data", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		data := SessionData{
			ID:            "session-123",
			Data:          map[string]any{"user_id": "456"},
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(1 * time.Hour),
			UpdatedAt:     time.Now(),
			Authenticated: true,
			UserID:        "user-456",
		}

		err := store.Set(ctx, data)

		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "session-123")
		require.NoError(t, err)
		assert.Equal(t, data.ID, retrieved.ID)
		assert.Equal(t, data.UserID, retrieved.UserID)
		assert.Equal(t, data.Authenticated, retrieved.Authenticated)
		assert.Equal(t, data.Data["user_id"], retrieved.Data["user_id"])
	})

	t.Run("should overwrite existing session", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		data1 := SessionData{
			ID:        "session-123",
			Data:      map[string]any{"version": 1},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data1)
		require.NoError(t, err)

		data2 := SessionData{
			ID:        "session-123",
			Data:      map[string]any{"version": 2},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(2 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err = store.Set(ctx, data2)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "session-123")
		require.NoError(t, err)
		assert.Equal(t, 2, retrieved.Data["version"])
	})

	t.Run("should create copy of data to prevent mutation", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		originalData := map[string]any{"key": "original"}
		data := SessionData{
			ID:        "session-123",
			Data:      originalData,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data)
		require.NoError(t, err)

		originalData["key"] = "modified"
		data.Data["another_key"] = "new"

		retrieved, err := store.Get(ctx, "session-123")
		require.NoError(t, err)


		assert.Equal(t, "original", retrieved.Data["key"])
		_, exists := retrieved.Data["another_key"]
		assert.False(t, exists, "New key should not appear in stored data")
	})
}

func TestMemoryStore_Get(t *testing.T) {
	t.Run("should retrieve existing session", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		data := SessionData{
			ID:            "session-456",
			Data:          map[string]any{"role": "admin"},
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(1 * time.Hour),
			UpdatedAt:     time.Now(),
			Authenticated: true,
			UserID:        "user-789",
		}

		store.Set(ctx, data)

		retrieved, err := store.Get(ctx, "session-456")

		require.NoError(t, err)
		assert.Equal(t, "session-456", retrieved.ID)
		assert.Equal(t, "admin", retrieved.Data["role"])
		assert.True(t, retrieved.Authenticated)
		assert.Equal(t, "user-789", retrieved.UserID)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		retrieved, err := store.Get(ctx, "non-existent")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.Empty(t, retrieved.ID)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		retrieved, err := store.Get(ctx, "")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.Empty(t, retrieved.ID)
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	t.Run("should delete existing session", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		data := SessionData{
			ID:        "session-to-delete",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		store.Set(ctx, data)

		_, err := store.Get(ctx, "session-to-delete")
		require.NoError(t, err)

		err = store.Delete(ctx, "session-to-delete")
		require.NoError(t, err)

		_, err = store.Get(ctx, "session-to-delete")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("should not error when deleting non-existent session", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		err := store.Delete(ctx, "non-existent")

		assert.NoError(t, err)
	})

	t.Run("should handle multiple deletes", func(t *testing.T) {
		store := NewMemoryStore()
		ctx := context.Background()

		data := SessionData{
			ID:        "session-123",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		store.Set(ctx, data)

		err := store.Delete(ctx, "session-123")
		assert.NoError(t, err)

		err = store.Delete(ctx, "session-123")
		assert.NoError(t, err)
	})
}
