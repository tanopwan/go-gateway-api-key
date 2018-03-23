package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var commands = []string{
	"DROP TABLE keys",
	"CREATE TABLE keys (id text, app_id text, key text, created timestamp without time zone)",
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
		fmt.Println("Exec command:", command)
		if _, err = db.Exec(command); err != nil {
			panic(err)
		}
	}

}
