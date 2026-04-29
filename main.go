package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type application struct {
	db *sql.DB
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		log.Fatalf("Database unreachable: %v", err)
		return nil, err
	}

	fmt.Println("Successfully connected with database!")

	return db, nil
}

func main() {
	dsn := "postgres://feedback:feedback@localhost/feedback?sslmode=disable"

	db, err := openDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &application{db: db}

	log.Println("Server running on :4000")
	http.ListenAndServe(":4000", app.routes())
}
