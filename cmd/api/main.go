package main

import (

	"github.com/AmiyoKm/go-backend/internal/db"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
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

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	//Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	}

	cfg := Config{
		Addr:   env.GetString("ADDR", ":8080"),
		DB:     DBConfig,
		Env:    env.GetString("ENVIRONMENT", "DEVELOPMENT"),
		ApiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
	}

	//Database
	db, err := db.New(cfg.DB.Addr, cfg.DB.maxOpenConns, cfg.DB.maxIdleConns, cfg.DB.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("DB connection pool established")

	store := store.NewStorage(db)

	app := &Application{
		Config: cfg,
		Store:  store,
		Logger: logger,
	}

	mux := app.mount()
	logger.Fatal(app.Run(mux))

}
