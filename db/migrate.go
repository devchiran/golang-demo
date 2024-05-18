package main

import (
	"flag"
	"fmt"
	"log"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	db   = flag.String("database", "postgres", "")
	host = flag.String("host", "localhost:5432", "")
	user = flag.String("user", "postgres", "")
	pass = flag.String("password", "", "")
)

func main() {
	flag.Parse()
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", *user, *pass, *host, *db)
	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}
