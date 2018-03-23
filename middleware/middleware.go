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

var (
	// TableName ... to keep api-keys for apps
	TableName = "api_keys"
)

// Service ...
type Service interface {
	GenerateAPIKey(appID string) (string, error)
	ValidateAPIKey(appID string, key string) error
}

type middleware struct {
	db        *sql.DB
	redisPool *redis.Pool
}

// NewService ... will return authService client that implements interfaces
func NewService(db *sql.DB, redisPool *redis.Pool) Service {
	return &middleware{db: db, redisPool: redisPool}
}

type appKey struct {
	ID      string
	AppID   string
	Key     string
	Created time.Time
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

	command := fmt.Sprintf("INSERT INTO %s(ID, APP_ID, KEY, CREATED) VALUES($1, $2, $3, $4)", TableName)
	_, err = m.db.Exec(command, ID, appID, key, time.Now())
	if pqErr, ok := err.(*pq.Error); ok {
		return "", fmt.Errorf("failed to insert with reason: %s:%s", pqErr.Code, pqErr.Message)
	}

	return key, nil
}

func (m *middleware) ValidateAPIKey(appID string, key string) error {

	appKey := appKey{}

	command := fmt.Sprintf("SELECT * FROM %s WHERE APP_ID = $1 AND KEY = $2", TableName)
	err := m.db.QueryRow(command, appID, key).Scan(&appKey.ID, &appKey.AppID, &appKey.Key, &appKey.Created)
	if err == sql.ErrNoRows {
		return fmt.Errorf("key is invalid")
	}

	if pqErr, ok := err.(*pq.Error); ok {
		return fmt.Errorf("failed to query api key with reason: %s : %s", pqErr.Code, pqErr.Message)
	}

	if err != nil {
		return fmt.Errorf("failed to query api key with error: %s", err.Error())
	}

	if appID != appKey.AppID || key != appKey.Key {
		return fmt.Errorf("failed to query api key")
	}

	fmt.Println("Validated Pass ", appKey.ID)
	return nil
}
