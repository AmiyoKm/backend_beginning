package main

import (
	"log"

	"github.com/AmiyoKm/go-backend/internal/db"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/joho/godotenv"
)

const version string = "0.0.1"

//	@title			SocialLink API
//	@description	API for SocialLink .
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	cfg := Config{
		Addr: env.GetString("ADDR", ":8080"),
		DB:   DBConfig,
		Env:  env.GetString("ENVIRONMENT", "DEVELOPMENT"),
		ApiURL: env.GetString("EXTERNAL_URL" , "localhost:8080"),
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
