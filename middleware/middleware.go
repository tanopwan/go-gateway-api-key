package middleware

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/lib/pq"
)

// Service ...
type Service interface {
	GenerateAPIKey(appID string) (string, error)
}

type middleware struct {
	db        *sql.DB
	redisPool *redis.Pool
}

// NewService ... will return authService client that implements interfaces
func NewService(db *sql.DB, redisPool *redis.Pool) Service {
	return &middleware{db: db, redisPool: redisPool}
}

// GenerateAPIKey ... will generate random api key for an app
func (m *middleware) GenerateAPIKey(appID string) (string, error) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	ID := hex.EncodeToString(b)

	b = make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		panic(err)
	}
	key := hex.EncodeToString(b)

	_, err = m.db.Exec("INSERT INTO keys(ID, APP_ID, KEY, CREATED) VALUES($1, $2, $3, $4)", ID, appID, key, time.Now())
	if pqErr, ok := err.(*pq.Error); ok {
		return "", fmt.Errorf("failed to insert with reason: %s:%s", pqErr.Code, pqErr.Message)
	}

	return key, nil
}
