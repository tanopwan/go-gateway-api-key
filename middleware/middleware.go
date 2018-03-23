package middleware

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

var (
	// TableName ... to keep api-keys for apps
	TableName = "api_keys"
)

// AppInfo ... is returned when validate api key
type AppInfo struct {
	AppID   string
	IsCache bool
}

// Service ...
type Service interface {
	GenerateAPIKey(appID string) (string, error)
	ValidateAPIKey(key string) (*AppInfo, error)
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

// ValidateAPIKeyMiddleWare ...
func ValidateAPIKeyMiddleWare(db *sql.DB, redisPool *redis.Pool) gin.HandlerFunc {
	middleware := NewService(db, redisPool)

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Api-Key")

		appInfo, err := middleware.ValidateAPIKey(apiKey)
		if err != nil {
			fmt.Println("err:", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "fail",
				"message": "forbidden",
			})
			return
		}

		c.Next()
		fmt.Println("appId:", appInfo.AppID)
	}
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

func (m *middleware) ValidateAPIKey(key string) (*AppInfo, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("parameter is invalid")
	}

	rd := m.redisPool.Get()
	defer rd.Close()

	cacheAppID, _ := redis.String(rd.Do("GET", key))
	if cacheAppID != "" {
		fmt.Println("Validated Pass found cache for appID", cacheAppID)
		return &AppInfo{AppID: cacheAppID, IsCache: true}, nil
	}

	appKey := appKey{}

	command := fmt.Sprintf("SELECT * FROM %s WHERE KEY = $1 ORDER BY CREATED DESC", TableName)
	err := m.db.QueryRow(command, key).Scan(&appKey.ID, &appKey.AppID, &appKey.Key, &appKey.Created)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("key is invalid")
	}

	if pqErr, ok := err.(*pq.Error); ok {
		return nil, fmt.Errorf("failed to query api key with reason: %s : %s", pqErr.Code, pqErr.Message)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query api key with error: %s", err.Error())
	}

	if appKey.AppID == "" {
		return nil, fmt.Errorf("failed to query api key")
	}

	fmt.Println("Validated Pass ", appKey.ID)

	_, err = rd.Do("SETEX", key, int64(time.Hour/time.Second), appKey.AppID)
	if err != nil {
		panic(err)
	}

	return &AppInfo{AppID: appKey.AppID, IsCache: false}, nil
}
