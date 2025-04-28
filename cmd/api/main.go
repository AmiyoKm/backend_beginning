package main

import (
	"log"

	"github.com/AmiyoKm/go-backend/internal/db"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/joho/godotenv"
)

const version string = "0.0.1"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	cfg := Config{
		Addr: env.GetString("ADDR", ":8080"),
		DB:   DBConfig,
		Env:  env.GetString("ENVIRONMENT", "DEVELOPMENT"),
	}

	db, err := db.New(cfg.DB.Addr, cfg.DB.maxOpenConns, cfg.DB.maxIdleConns, cfg.DB.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	log.Println("DB connection pool established")

	store := store.NewStorage(db)

	app := &Application{
		Config: cfg,
		Store:  store,
	}

	mux := app.mount()
	log.Fatal(app.Run(mux))

}
