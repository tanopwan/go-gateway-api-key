package test

import (
	"database/sql"
	"testing"

	"github.com/tanopwan/go-gateway-api-key/middleware"
)

func setup() *sql.DB {
	db, err := sql.Open("postgres", "postgres://app:password@localhost/tanopwan?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

func shutdown() {

}

func TestCreateAPIKey(t *testing.T) {
	db := setup()
	defer db.Close()

	m := middleware.NewService(db, nil)

	key, err := m.GenerateAPIKey("1234")
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}
	t.Log(key)
}

func TestValidateAPIKey(t *testing.T) {
	db := setup()
	defer db.Close()

	m := middleware.NewService(db, nil)

	key, err := m.GenerateAPIKey("1234")
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}
	t.Log(key)

	err = m.ValidateAPIKey("1234", key)
	if err != nil {
		t.Log(err.Error())
		panic(err)
	}
}
