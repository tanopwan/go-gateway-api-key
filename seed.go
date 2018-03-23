package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/tanopwan/go-gateway-api-key/middleware"
)

var commands = []string{
	"DROP TABLE %s",
	"CREATE TABLE %s (id text, app_id text, key text, created timestamp without time zone)",
}

func main() {

	db, err := sql.Open("postgres", "postgres://app:password@localhost/tanopwan?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	for _, command := range commands {
		format := fmt.Sprintf(command, middleware.TableName)
		fmt.Println("Exec command:", format)
		if _, err = db.Exec(format); err != nil {
			panic(err)
		}
	}

}
