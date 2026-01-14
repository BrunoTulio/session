package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BrunoTulio/session"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	prefix string
	client *redis.Client
}

func NewStore(client *redis.Client, prefix string) session.Store {
	return &Store{
		prefix: prefix,
		client: client,
	}
}

func (s *Store) Get(ctx context.Context, id string) (session.SessionData, error) {
	key := s.prefix + id

	data, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return session.SessionData{}, session.ErrSessionNotFound
	}

	if err != nil {
		return session.SessionData{}, fmt.Errorf("redis get failed: %w", err)
	}

	var sess session.SessionData
	if err := json.Unmarshal(data, &sess); err != nil {
		return session.SessionData{}, fmt.Errorf("unmarshal failed: %w", err)
	}

	return sess, nil
}

func (s *Store) Set(ctx context.Context, session session.SessionData) error {
	key := s.prefix + session.ID

	data, err := json.Marshal(&session)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl < 0 {
		ttl = time.Second
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Store) Delete(ctx context.Context, id string) error {
	key := s.prefix + id
	return s.client.Del(ctx, key).Err()
}
