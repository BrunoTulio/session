// pkg/session/redis/store_test.go

package redis_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	redisstore "github.com/BrunoTulio/session/redis"

	"github.com/BrunoTulio/session"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestNewStore(t *testing.T) {
	t.Run("should create new redis store", func(t *testing.T) {
		client, _ := setupRedis(t)

		store := redisstore.NewStore(client, "session:")

		assert.NotNil(t, store)
	})
}

func TestRedisStore_Set(t *testing.T) {
	t.Run("should store session data", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
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

		key := "session:session-123"
		assert.True(t, mr.Exists(key))

		storedJSON, err := mr.Get(key)
		require.NoError(t, err)

		var stored session.SessionData
		err = json.Unmarshal([]byte(storedJSON), &stored)
		require.NoError(t, err)

		assert.Equal(t, data.ID, stored.ID)
		assert.Equal(t, data.UserID, stored.UserID)
		assert.Equal(t, data.Authenticated, stored.Authenticated)
	})

	t.Run("should set TTL based on ExpiresAt", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "session-ttl",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data)
		require.NoError(t, err)

		ttl := mr.TTL("session:session-ttl")

		assert.Greater(t, ttl.Seconds(), 3590.0) // pelo menos 59:50
		assert.Less(t, ttl.Seconds(), 3610.0)    // no máximo 1:00:10
	})

	t.Run("should set minimum TTL for expired session", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "session-expired",
			Data:      make(map[string]any),
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Já expirou!
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data)
		require.NoError(t, err)

		ttl := mr.TTL("session:session-expired")
		assert.Greater(t, ttl.Seconds(), 0.0)
	})

	t.Run("should overwrite existing session", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data1 := session.SessionData{
			ID:        "session-123",
			Data:      map[string]any{"version": 1},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}
		store.Set(ctx, data1)

		data2 := session.SessionData{
			ID:        "session-123",
			Data:      map[string]any{"version": 2},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(2 * time.Hour),
			UpdatedAt: time.Now(),
		}
		err := store.Set(ctx, data2)
		require.NoError(t, err)

		storedJSON, _ := mr.Get("session:session-123")
		var stored session.SessionData
		json.Unmarshal([]byte(storedJSON), &stored)

		assert.Equal(t, 2, int(stored.Data["version"].(float64)))
	})

	t.Run("should use custom prefix", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "custom:prefix:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "test",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		store.Set(ctx, data)

		assert.True(t, mr.Exists("custom:prefix:test"))
		assert.False(t, mr.Exists("session:test"))
	})
}

func TestRedisStore_Get(t *testing.T) {
	t.Run("should retrieve existing session", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:            "session-456",
			Data:          map[string]any{"role": "admin"},
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(1 * time.Hour),
			UpdatedAt:     time.Now(),
			Authenticated: true,
			UserID:        "user-789",
		}

		jsonData, _ := json.Marshal(data)
		mr.Set("session:session-456", string(jsonData))

		retrieved, err := store.Get(ctx, "session-456")

		require.NoError(t, err)
		assert.Equal(t, "session-456", retrieved.ID)
		assert.Equal(t, "admin", retrieved.Data["role"])
		assert.True(t, retrieved.Authenticated)
		assert.Equal(t, "user-789", retrieved.UserID)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		retrieved, err := store.Get(ctx, "non-existent")

		assert.Error(t, err)
		assert.ErrorIs(t, err, session.ErrSessionNotFound)
		assert.Empty(t, retrieved.ID)
	})

	t.Run("should handle expired session", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "session-expire",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Millisecond),
			UpdatedAt: time.Now(),
		}

		jsonData, _ := json.Marshal(data)
		mr.Set("session:session-expire", string(jsonData))
		mr.SetTTL("session:session-expire", 1*time.Millisecond)

		mr.FastForward(10 * time.Millisecond)

		_, err := store.Get(ctx, "session-expire")

		assert.Error(t, err)
		assert.ErrorIs(t, err, session.ErrSessionNotFound)
	})

	t.Run("should return error for invalid JSON", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		mr.Set("session:invalid", "{ invalid json }")

		_, err := store.Get(ctx, "invalid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal failed")
	})

	t.Run("should use custom prefix", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "app:sessions:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "test",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		jsonData, _ := json.Marshal(data)
		mr.Set("app:sessions:test", string(jsonData))

		retrieved, err := store.Get(ctx, "test")

		require.NoError(t, err)
		assert.Equal(t, "test", retrieved.ID)
	})
}

func TestRedisStore_Delete(t *testing.T) {
	t.Run("should delete existing session", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "session-delete",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		store.Set(ctx, data)
		assert.True(t, mr.Exists("session:session-delete"))

		err := store.Delete(ctx, "session-delete")
		require.NoError(t, err)

		assert.False(t, mr.Exists("session:session-delete"))
	})

	t.Run("should not error when deleting non-existent session", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		err := store.Delete(ctx, "non-existent")

		assert.NoError(t, err)
	})

	t.Run("should handle multiple deletes", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
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

	t.Run("should use custom prefix", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "my:prefix:")
		ctx := context.Background()

		jsonData, _ := json.Marshal(session.SessionData{ID: "test"})
		mr.Set("my:prefix:test", string(jsonData))

		err := store.Delete(ctx, "test")
		require.NoError(t, err)

		assert.False(t, mr.Exists("my:prefix:test"))
	})
}

func TestRedisStore_Integration(t *testing.T) {
	t.Run("should handle complete CRUD cycle", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:            "integration-test",
			Data:          map[string]any{"key": "value"},
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(1 * time.Hour),
			UpdatedAt:     time.Now(),
			Authenticated: false,
		}

		err := store.Set(ctx, data)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "integration-test")
		require.NoError(t, err)
		assert.Equal(t, data.ID, retrieved.ID)
		assert.Equal(t, "value", retrieved.Data["key"])

		retrieved.Data["key"] = "updated"
		retrieved.Authenticated = true
		err = store.Set(ctx, retrieved)
		require.NoError(t, err)

		updated, err := store.Get(ctx, "integration-test")
		require.NoError(t, err)
		assert.Equal(t, "updated", updated.Data["key"])
		assert.True(t, updated.Authenticated)

		err = store.Delete(ctx, "integration-test")
		require.NoError(t, err)

		_, err = store.Get(ctx, "integration-test")
		assert.ErrorIs(t, err, session.ErrSessionNotFound)
	})
}

func TestRedisStore_EdgeCases(t *testing.T) {
	t.Run("should handle empty prefix", func(t *testing.T) {
		client, mr := setupRedis(t)
		store := redisstore.NewStore(client, "")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "test",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		store.Set(ctx, data)

		assert.True(t, mr.Exists("test"))
	})

	t.Run("should handle nil Data map", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "nil-data",
			Data:      nil,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "nil-data")
		require.NoError(t, err)
		assert.Nil(t, retrieved.Data)
	})

	t.Run("should handle empty Data map", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")
		ctx := context.Background()

		data := session.SessionData{
			ID:        "empty-data",
			Data:      map[string]any{},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "empty-data")
		require.NoError(t, err)
		assert.NotNil(t, retrieved.Data)
		assert.Len(t, retrieved.Data, 0)
	})
}

func TestRedisStore_ContextCancellation(t *testing.T) {
	t.Run("should respect context cancellation on Get", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := store.Get(ctx, "test")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("should respect context timeout on Set", func(t *testing.T) {
		client, _ := setupRedis(t)
		store := redisstore.NewStore(client, "session:")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond)

		data := session.SessionData{
			ID:        "test",
			Data:      make(map[string]any),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}

		err := store.Set(ctx, data)

		assert.Error(t, err)
	})
}
