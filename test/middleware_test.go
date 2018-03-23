package test

import (
	"database/sql"
	"testing"

	"github.com/tanopwan/go-gateway-api-key/middleware"
)

func TestCreateAPIKey(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://app:password@localhost/tanopwan?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	m := middleware.NewService(db, nil)

	key, err := m.GenerateAPIKey("1234")
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}
	t.Log(key)
}
