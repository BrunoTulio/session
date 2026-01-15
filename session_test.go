package session

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_New(t *testing.T) {
	t.Run("should new session modified", func(t *testing.T) {
		s := NewSession(time.Hour * 24)

		assert.NotNil(t, s)
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())
	})

	t.Run("should create new session unique ids", func(t *testing.T) {
		ids := make(map[string]bool)
		interactions := 1000
		for range interactions {
			s := NewSession(time.Hour * 24)

			assert.NotNil(t, s)

			if ids[s.ID] {
				assert.FailNow(t, "duplicate session id")
			}

			ids[s.ID] = true
		}
		assert.Equal(t, interactions, len(ids))
	})
}

func TestSession_GetSessionData(t *testing.T) {
	t.Run("should get session data", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		s.Set("value", "value")
		s.Authenticate("123")

		assert.NotNil(t, s)

		data := s.GetSessionData()
		assert.NotNil(t, data)
		assert.True(t, data.Authenticated)
		val, ok := data.Get("value")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})
}

func TestSession_NewFromData(t *testing.T) {
	t.Run("should new session from data", func(t *testing.T) {
		data := NewSessionData(time.Hour * 24)
		assert.NotNil(t, data)

		s := NewSessionFromData(data)
		assert.NotNil(t, s)
		assert.NotNil(t, s.Data)
	})

	t.Run("should new session from data and make data", func(t *testing.T) {
		data := SessionData{
			Data: nil,
		}
		assert.NotNil(t, data)
		assert.Nil(t, data.Data)

		s := NewSessionFromData(data)
		assert.NotNil(t, s)
		assert.NotNil(t, s.SessionData.Data)
	})
}

func TestSession_GetAndSet(t *testing.T) {
	t.Run("should get and set session", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		assert.NotNil(t, s)

		s.Set("user_id", "123")
		s.Set("role", "admin")
		s.Set("count", 42)

		val, ok := s.Get("user_id")
		assert.True(t, ok)
		assert.Equal(t, "123", val)

		val, ok = s.Get("role")
		assert.True(t, ok)
		assert.Equal(t, "admin", val)

		val, ok = s.Get("count")
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("should return false there is no key", func(t *testing.T) {
		data := NewSession(time.Hour * 24)
		val, ok := data.Get("user_id")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("should update UpdatedAt on Set", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }
		data := NewSession(time.Hour * 24)

		assert.Equal(t, timeBeforeUpdate, data.UpdatedAt)

		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }
		data.Set("user_id", "123")
		now = func() time.Time { return time.Now() }

		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})
}

func TestSession_Delete(t *testing.T) {
	t.Run("should delete success", func(t *testing.T) {
		data := NewSession(time.Hour * 24)

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

	t.Run("should update modified on Delete", func(t *testing.T) {
		data := NewSession(time.Hour * 24)
		assert.True(t, data.modified)
		assert.True(t, data.IsModified())

		data.MarkClean()
		assert.False(t, data.modified)
		assert.False(t, data.IsModified())

		data.Set("user_id", "123")
		val, ok := data.Get("user_id")
		assert.True(t, ok)
		assert.NotNil(t, val)
		assert.Equal(t, "123", val)
		assert.True(t, data.modified)
		assert.True(t, data.IsModified())
	})

	t.Run("should update UpdatedAt on Delete", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }
		data := NewSession(time.Hour * 24)
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

func TestSession_Authenticate(t *testing.T) {
	t.Run("should authenticate", func(t *testing.T) {
		data := NewSession(time.Hour * 24)

		assert.False(t, data.Authenticated)
		assert.Empty(t, data.UserID)

		data.Authenticate("1234")
		assert.True(t, data.Authenticated)
		assert.NotEmpty(t, data.UserID)
		assert.Equal(t, "1234", data.UserID)
	})

	t.Run("should update UpdatedAt on Authenticate", func(t *testing.T) {
		timeBeforeUpdate := time.Now().AddDate(-1, 0, 0)
		now = func() time.Time { return timeBeforeUpdate }

		data := NewSession(time.Hour * 24)

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

	t.Run("should update modified on Authenticate", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())

		s.MarkClean()
		assert.False(t, s.modified)
		assert.False(t, s.IsModified())

		s.Authenticate("1234")
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())
	})
}

func TestSession_Unauthenticate(t *testing.T) {
	t.Run("should unauthenticate", func(t *testing.T) {
		data := NewSession(time.Hour * 24)
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

		data := NewSession(time.Hour * 24)
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

	t.Run("should update modified on Unauthenticate", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())

		s.MarkClean()
		assert.False(t, s.modified)
		assert.False(t, s.IsModified())

		s.Authenticate("1234")
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())

		s.MarkClean()
		assert.False(t, s.modified)
		assert.False(t, s.IsModified())

		s.Unauthenticate()
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())
	})
}

func TestSession_Renew(t *testing.T) {
	t.Run("should renew", func(t *testing.T) {
		timeDefault := time.Now().AddDate(0, 0, -1)
		now = func() time.Time { return timeDefault }
		data := NewSession(time.Hour * 24)
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

		data := NewSession(time.Hour * 24)
		assert.Equal(t, timeBeforeUpdate, data.UpdatedAt)
		timeAfterUpdate := time.Now()
		now = func() time.Time { return timeAfterUpdate }

		data.Renew(time.Hour * 24)
		assert.Equal(t, timeAfterUpdate, data.UpdatedAt)
	})

	t.Run("should update modified on Renew", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())

		s.MarkClean()
		assert.False(t, s.modified)
		assert.False(t, s.IsModified())

		s.Renew(time.Hour * 24)
		assert.True(t, s.modified)
		assert.True(t, s.IsModified())
	})
}

func TestSession_IsExpired(t *testing.T) {
	t.Run("should return true expired", func(t *testing.T) {
		now = func() time.Time { return time.Now().AddDate(-1, 0, 0) }
		data := NewSession(time.Hour * 24)
		now = func() time.Time { return time.Now() }

		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.True(t, isExpired, "should be expired")
	})

	t.Run("should return false not expired", func(t *testing.T) {
		data := NewSession(time.Hour * 24)

		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.False(t, isExpired, "should be expired")
	})

	t.Run("should return true it expires at the exact time", func(t *testing.T) {
		now = func() time.Time { return time.Now().Add(-24 * time.Hour) }
		data := NewSession(time.Hour * 24)
		now = func() time.Time { return time.Now() }

		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.True(t, isExpired, "should be expired")
	})

	t.Run("should return false with one minute left to expire", func(t *testing.T) {
		now = func() time.Time { return time.Now().Add(-(23*time.Hour + 9*time.Minute)) }
		data := NewSession(time.Hour * 24)
		now = func() time.Time { return time.Now() }

		assert.NotNil(t, data)
		isExpired := data.IsExpired()

		assert.False(t, isExpired, "should be expired")
	})
}

func TestSession_MarkClean(t *testing.T) {
	t.Run("should mark clean", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		assert.NotNil(t, s)

		assert.True(t, s.IsModified())
		assert.True(t, s.modified)

		s.MarkClean()
		assert.False(t, s.IsModified())
		assert.False(t, s.modified)
	})
}

func TestSession_IsAuthenticated(t *testing.T) {
	t.Run("should return true authenticated", func(t *testing.T) {
		s := NewSession(time.Hour * 24)
		assert.NotNil(t, s)

		assert.False(t, s.IsAuthenticated())
		s.Authenticate("123")
		assert.True(t, s.IsAuthenticated())
	})
}

func TestSession_Concurrent(t *testing.T) {
	t.Run("should handle concurrent writes without data loss", func(t *testing.T) {
		sess := NewSession(1 * time.Hour)

		var wg sync.WaitGroup
		goroutines := 10
		writesPerGoroutine := 100

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < writesPerGoroutine; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j)
					sess.Set(key, fmt.Sprintf("value_%d_%d", id, j))
				}
			}(i)
		}

		wg.Wait()

		totalExpected := goroutines * writesPerGoroutine

		data := sess.GetSessionData()
		actualCount := len(data.Data)

		assert.Equal(t, totalExpected, actualCount,
			"Expected %d keys, got %d - possible race condition!",
			totalExpected, actualCount)

		for i := 0; i < goroutines; i++ {
			for j := 0; j < writesPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", i, j)
				expectedValue := fmt.Sprintf("value_%d_%d", i, j)

				val, ok := sess.Get(key)
				assert.True(t, ok, "Key %s should exist", key)
				assert.Equal(t, expectedValue, val, "Value mismatch for key %s", key)
			}
		}
	})

	t.Run("should handle concurrent reads and writes", func(t *testing.T) {
		sess := NewSession(1 * time.Hour)

		for i := 0; i < 50; i++ {
			sess.Set(fmt.Sprintf("key_%d", i), i)
		}
		sess.MarkClean()

		var wg sync.WaitGroup
		iterations := 1000
		errors := make(chan error, iterations*2)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("write_key_%d", i)
				sess.Set(key, i)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("key_%d", i%50)
				val, ok := sess.Get(key)
				if ok {
					if val != i%50 {
						errors <- fmt.Errorf("expected %d, got %v for key %s", i%50, val, key)
					}
				}
			}
		}()

		wg.Wait()
		close(errors)

		errorCount := 0
		for err := range errors {
			t.Error(err)
			errorCount++
		}

		assert.Equal(t, 0, errorCount, "Should have no read errors")
	})

	t.Run("should correctly track modified flag under concurrency", func(t *testing.T) {
		sess := NewSession(1 * time.Hour)
		sess.MarkClean()

		var wg sync.WaitGroup
		goroutines := 100

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				sess.Set(fmt.Sprintf("key_%d", id), id)
			}(i)
		}

		wg.Wait()

		assert.True(t, sess.IsModified(),
			"Session should be marked as modified after concurrent writes")
	})
}

func TestSession_GenerateId(t *testing.T) {
	t.Run("should generate id", func(t *testing.T) {
		ids := make(map[string]bool)
		iterations := 1000
		for range iterations {
			id := generateId()
			if _, ok := ids[id]; ok {
				assert.Fail(t, "duplicate id")
			}
			ids[id] = true
		}
		assert.Len(t, ids, iterations)
	})
}
