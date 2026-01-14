package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionData_New(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data, err := NewSessionData(1 * time.Hour)
		assert.NoError(t, err)
		assert.NotNil(t, data)
	})
}

func TestSessionData_UniqueIds(t *testing.T) {
	t.Run("should generate unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			data, err := NewSessionData(1 * time.Hour)
			assert.NoError(t, err)

			if ids[data.ID] {
				t.Fatalf("Duplicate ID found: %s", data.ID)
			}
			ids[data.ID] = true
		}

		assert.Len(t, ids, iterations)
	})
}

func TestSessionData_GetAndSet(t *testing.T) {
	t.Run("should set and get data", func(t *testing.T) {
		data, _ := NewSessionData(time.Hour * 24)

		data.Set("user_id", "123")
		data.Set("role", "admin")
		data.Set("count", 42)

		val, ok := data.Get("user_id")
		assert.True(t, ok)
		assert.Equal(t, "123", val)

		val, ok = data.Get("role")
		assert.True(t, ok)
		assert.Equal(t, "admin", val)

		val, ok = data.Get("count")
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("should return false there is no key", func(t *testing.T) {
		data, _ := NewSessionData(time.Hour * 24)
		val, ok := data.Get("user_id")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("should update UpdatedAt on Set", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }
		data, _ := NewSessionData(time.Hour * 24)

		assert.Equal(t, timeBeforeUpdate, data.UpdatedAt)

		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }
		data.Set("user_id", "123")
		now = func() time.Time { return time.Now() }

		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})
}

func TestSessionData_Delete(t *testing.T) {
	t.Run("should delete success", func(t *testing.T) {
		data, _ := NewSessionData(time.Hour * 24)

		data.Set("user_id", "123")
		val, ok := data.Get("user_id")
		assert.True(t, ok)
		assert.NotNil(t, val)
		assert.Equal(t, "123", val)

		data.Delete("user_id")

		val, ok = data.Get("user_id")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("should update UpdatedAt on Delete", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }
		data, _ := NewSessionData(time.Hour * 24)
		assert.Equal(t, timeBeforeUpdate, data.UpdatedAt)
		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }
		data.Set("user_id", "123")
		val, ok := data.Get("user_id")
		assert.True(t, ok)
		assert.NotNil(t, val)
		assert.Equal(t, "123", val)
		data.Delete("user_id")

		now = func() time.Time { return time.Now() }
		val, ok = data.Get("user_id")
		assert.False(t, ok)
		assert.Nil(t, val)
		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})
}

func TestSessionData_Authenticate(t *testing.T) {
	t.Run("should authenticate", func(t *testing.T) {
		data, _ := NewSessionData(time.Hour * 24)

		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)

		data.Authenticate("1234")
		assert.True(t, data.Authenticated)
		assert.NotEmpty(t, data.UserID)
		assert.Equal(t, "1234", data.UserID)
	})

	t.Run("should update UpdatedAt on Delete", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }

		data, _ := NewSessionData(time.Hour * 24)

		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)
		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }

		data.Authenticate("1234")
		now = func() time.Time { return time.Now() }
		assert.True(t, data.Authenticated)
		assert.NotEmpty(t, data.UserID)
		assert.Equal(t, "1234", data.UserID)
		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})
}

func TestSessionData_Unauthenticate(t *testing.T) {
	t.Run("should unauthenticate", func(t *testing.T) {
		data, _ := NewSessionData(time.Hour * 24)
		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)

		data.Authenticate("1234")
		assert.True(t, data.Authenticated)
		assert.NotEmpty(t, data.UserID)
		assert.Equal(t, "1234", data.UserID)

		data.Unauthenticate()
		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)
	})

	t.Run("should update UpdatedAt on Unauthenticate", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }

		data, _ := NewSessionData(time.Hour * 24)
		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)
		assert.Equal(t, timeBeforeUpdate, data.UpdatedAt)

		data.Authenticate("1234")
		assert.True(t, data.Authenticated)
		assert.NotEmpty(t, data.UserID)
		assert.Equal(t, "1234", data.UserID)

		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }

		data.Unauthenticate()
		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)
		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})
}

func TestSessionData_Renew(t *testing.T) {
	t.Run("should renew", func(t *testing.T) {
		timeDefault := time.Now().AddDate(0, 0, -1)
		now = func() time.Time { return timeDefault }
		data, _ := NewSessionData(time.Hour * 24)
		expiredExpected := timeDefault.Add(time.Hour * 24)
		expiredOld := data.ExpiresAt

		assert.Equal(t, expiredExpected, data.ExpiresAt)

		timeDefault = time.Now()
		now = func() time.Time { return timeDefault }

		data.Renew(time.Hour * 24)
		expiredExpected = timeDefault.Add(time.Hour * 24)

		assert.Equal(t, expiredExpected, data.ExpiresAt)
		assert.True(t, data.ExpiresAt.After(expiredOld))
	})

	t.Run("should update UpdatedAt on Renew", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }

		data, _ := NewSessionData(time.Hour * 24)
		assert.Equal(t, timeBeforeUpdate, data.UpdatedAt)
		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }

		data.Renew(time.Hour * 24)
		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})
}

func TestSessionData_IsExpired(t *testing.T) {
	t.Run("should return true expired", func(t *testing.T) {
		now = func() time.Time { return time.Now().AddDate(-1, 0, 0) }
		data, err := NewSessionData(time.Hour * 24)
		now = func() time.Time { return time.Now() }

		assert.NoError(t, err)
		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.True(t, isExpired, "should be expired")
	})

	t.Run("should return false not expired", func(t *testing.T) {
		data, err := NewSessionData(time.Hour * 24)

		assert.NoError(t, err)
		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.False(t, isExpired, "should be expired")
	})

	t.Run("should return true it expires at the exact time", func(t *testing.T) {
		now = func() time.Time { return time.Now().Add(-24 * time.Hour) }
		data, err := NewSessionData(time.Hour * 24)
		now = func() time.Time { return time.Now() }

		assert.NoError(t, err)
		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.True(t, isExpired, "should be expired")
	})

	t.Run("should return false with one minute left to expire", func(t *testing.T) {
		now = func() time.Time { return time.Now().Add(-(23*time.Hour + 9*time.Minute)) }
		data, err := NewSessionData(time.Hour * 24)
		now = func() time.Time { return time.Now() }

		assert.NoError(t, err)
		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.False(t, isExpired, "should be expired")
	})
}
