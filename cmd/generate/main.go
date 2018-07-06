package main

import (
	"fmt"
	"time"

	"bitbucket.org/morrocio/kafra/conn"
	"github.com/garyburd/redigo/redis"
	"github.com/tanopwan/go-gateway-api-key/middleware"
)

func main() {
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

	db, err := conn.GetNewPGCon(1, 1)
	if err != nil {
		panic(err)
	}

	service := middleware.NewService(db, redisPool)
	key, err := service.GenerateAPIKey("ov0EfK1OdeBSNIHnTv0uqA")
	if err != nil {
		panic(err)
	}
	fmt.Printf("key: %s\n", key)
}
