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
	ValidateAPIKey(appID string, key string) (bool, error)
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

func (m *middleware) ValidateAPIKey(appID string, key string) (bool, error) {
	if len(appID) < 16 || len(key) < 64 {
		return false, fmt.Errorf("parameters are invalid")
	}

	rd := m.redisPool.Get()
	defer rd.Close()

	cacheAppID, _ := redis.String(rd.Do("GET", key))
	if cacheAppID == appID {
		fmt.Println("Validated Pass found cache for appID", appID)
		return true, nil
	}

	appKey := appKey{}

	command := fmt.Sprintf("SELECT * FROM %s WHERE APP_ID = $1 AND KEY = $2", TableName)
	err := m.db.QueryRow(command, appID, key).Scan(&appKey.ID, &appKey.AppID, &appKey.Key, &appKey.Created)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("key is invalid")
	}

	if pqErr, ok := err.(*pq.Error); ok {
		return false, fmt.Errorf("failed to query api key with reason: %s : %s", pqErr.Code, pqErr.Message)
	}

	if err != nil {
		return false, fmt.Errorf("failed to query api key with error: %s", err.Error())
	}

	if appID != appKey.AppID || key != appKey.Key {
		return false, fmt.Errorf("failed to query api key")
	}

	fmt.Println("Validated Pass ", appKey.ID)

	_, err = rd.Do("SETEX", key, int64(time.Hour/time.Second), appID)
	if err != nil {
		panic(err)
	}

	return false, nil
}
