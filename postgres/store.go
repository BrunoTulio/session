package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BrunoTulio/session"
)

// Store implements session.Store interface using PostgreSQL.
//
// IMPORTANT: Before using this store, you must create the sessions table.
// Run the migration SQL from migrations/sessions.sql or use the CreateSessionsTable function.
//
// Required table schema:
//
//	CREATE TABLE IF NOT EXISTS sessions (
//	    id VARCHAR(48) PRIMARY KEY,
//	    user_id VARCHAR(255),
//	    authenticated BOOLEAN NOT NULL DEFAULT FALSE,
//	    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
//	    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
//	    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
//	    data JSONB NOT NULL DEFAULT '{}'::jsonb
//	);
//
//	CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
//	CREATE INDEX idx_sessions_user_id ON sessions(user_id) WHERE user_id IS NOT NULL;
//	CREATE INDEX idx_sessions_authenticated ON sessions(authenticated) WHERE authenticated = TRUE;
type Store struct {
	db              *sql.DB
	log             session.Logger
	cleanerInterval time.Duration
}

func (s *Store) Get(ctx context.Context, id string) (session.SessionData, error) {
	const query = `  SELECT id, 
           user_id,
          authenticated,
    data ,
    expires_at ,
    created_at ,
    updated_at 
       FROM sessions
       WHERE id = $1 AND expires_at > NOW()`

	var (
		sessionIDRow     string
		userIDRow        sql.NullString
		authenticatedRow bool
		dataJSON         []byte
		expiresAtRow     time.Time
		createdAtRow     time.Time
		updatedAtRow     time.Time
	)

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&sessionIDRow,
		&userIDRow,
		&authenticatedRow,
		&dataJSON,
		&expiresAtRow,
		&createdAtRow,
		&updatedAtRow,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return session.SessionData{}, session.ErrSessionNotFound
	}

	if err != nil {
		return session.SessionData{}, fmt.Errorf("postgres get failed: %w", err)
	}

	var data map[string]any
	if len(dataJSON) > 0 {
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			return session.SessionData{}, fmt.Errorf("unmarshal data failed: %w", err)
		}
	} else {
		data = make(map[string]any)
	}

	sess := session.SessionData{
		ID:            sessionIDRow,
		UserID:        userIDRow.String,
		Authenticated: authenticatedRow,
		Data:          data,
		ExpiresAt:     expiresAtRow,
		CreatedAt:     createdAtRow,
		UpdatedAt:     updatedAtRow,
	}

	return sess, nil
}

func (s *Store) Set(ctx context.Context, session session.SessionData) error {
	dataJSON, err := json.Marshal(session.Data)
	if err != nil {
		return fmt.Errorf("marshal data failed: %w", err)
	}

	const query = `
       INSERT INTO sessions (id, user_id, authenticated, data, expires_at, created_at, updated_at)
       VALUES ($1, $2, $3, $4, $5, $6, $7)
       ON CONFLICT (id) 
       DO UPDATE SET 
          user_id = EXCLUDED.user_id,
          authenticated = EXCLUDED.authenticated,
          data = EXCLUDED.data,
          expires_at = EXCLUDED.expires_at,
          updated_at = EXCLUDED.updated_at
    `

	userID := sql.NullString{
		String: session.UserID,
		Valid:  session.UserID != "",
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		session.ID,
		userID,
		session.Authenticated,
		dataJSON,
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert/update failed: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	const query = "DELETE FROM sessions WHERE id = $1"

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

func (s *Store) cleanExpiredSessions(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	const query = "DELETE FROM sessions WHERE expires_at < NOW()"

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		result, err := s.db.ExecContext(ctx, query)

		if err != nil {
			s.log.Errorf("Failed to clean expired sessions: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			if rows > 0 {
				s.log.Infof("Cleaned %d expired sessions", rows)
			}
		}

		cancel()
	}
}

func New(db *sql.DB, log session.Logger, cleanerInterval time.Duration) session.Store {
	s := &Store{
		db:              db,
		log:             log,
		cleanerInterval: cleanerInterval,
	}

	go s.cleanExpiredSessions(cleanerInterval)

	return s
}
