package test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/tanopwan/go-gateway-api-key/middleware"
)

func setup() (*sql.DB, *redis.Pool) {
	db, err := sql.Open("postgres", "postgres://app:password@localhost/tanopwan?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// init redis pool
	redisPool := &redis.Pool{
		MaxIdle:     2,
		IdleTimeout: 60 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) > time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	return db, redisPool
}

func shutdown(db *sql.DB, appID string, key string) {
	command := fmt.Sprintf("DELETE FROM %s WHERE APP_ID = $1 AND KEY = $2", middleware.TableName)
	_, err := db.Exec(command, appID, key)
	if err != nil {
		panic(err)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func TestCreateAPIKey(t *testing.T) {
	db, _ := setup()
	defer db.Close()

	m := middleware.NewService(db, nil)

	appID := "1234567812345678"
	key, err := m.GenerateAPIKey(appID)
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}
	assertEqual(t, len(key), 64, fmt.Sprintf("invalid key length of %d", len(key)))
	shutdown(db, appID, key)
}

func TestValidateAPIKeyFromDB(t *testing.T) {
	db, redisPool := setup()
	defer db.Close()

	m := middleware.NewService(db, redisPool)

	appID := "1234567812345678"
	key, err := m.GenerateAPIKey(appID)
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}

	isCache, err := m.ValidateAPIKey(appID, key)
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}

	assertEqual(t, isCache, false, "")

	rd := redisPool.Get()
	defer rd.Close()
	cacheAppID, _ := redis.String(rd.Do("GET", key))
	fmt.Println("cacheAppID: ", cacheAppID)
	assertEqual(t, cacheAppID, appID, "not found in cache")

	isCache, err = m.ValidateAPIKey(appID, key)
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}

	assertEqual(t, isCache, true, "")

	shutdown(db, appID, key)
}

func TestValidateAPIKeyFromRedis(t *testing.T) {
	db, redisPool := setup()
	defer db.Close()

	m := middleware.NewService(db, redisPool)

	rd := redisPool.Get()
	defer rd.Close()

	appID := "1234567812345678"
	key := "1234567812345678123456781234567812345678123456781234567812345678"
	_, err := rd.Do("SETEX", key, int64(time.Hour/time.Second), appID)
	if err != nil {
		panic(err)
	}

	isCache, err := m.ValidateAPIKey(appID, key)
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}

	assertEqual(t, isCache, true, "")
	shutdown(db, appID, key)
}
