package main

import (
	"log"

	"github.com/AmiyoKm/go-backend/internal/db"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable")
	conn, err := db.New(addr, 30, 30, "15m")
	if err != nil {
		log.Fatal(err)
	}
	store := store.NewStorage(conn)
	db.Seed(store)
}
