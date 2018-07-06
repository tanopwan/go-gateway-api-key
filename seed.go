package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/tanopwan/go-gateway-api-key/middleware"
)

var commands = []string{
	// "DROP TABLE %s",
	"CREATE TABLE %s (id text, app_id text, key text, created timestamp without time zone)",
}

func main() {

	host := os.Getenv("PG_HOST")
	database := os.Getenv("PG_DATABASE")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, database)
	db, err := sql.Open("postgres", connStr)
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
