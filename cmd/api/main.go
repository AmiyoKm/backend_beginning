package main

import (
	"time"

	"github.com/AmiyoKm/go-backend/internal/db"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/mailer"
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
	mailCgf := mailConfig{
		exp:       time.Hour * 24 * 3,
		fromEmail: env.GetString("FROM_EMAIL", "demomailtrap.com"),
		mailTrap: mailTrapConfig{
			apiKey: env.GetString("MAILTRAP_API_KEY", "fa17472ff57682f84f31cae401fd8556"),
		},
	}

	cfg := Config{
		Addr:        env.GetString("ADDR", ":8080"),
		DB:          DBConfig,
		Env:         env.GetString("ENVIRONMENT", "DEVELOPMENT"),
		ApiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		FrontendURL: env.GetString("FRONTEND_URL", " "),
		Mail:        mailCgf,
	}

	//Database
	db, err := db.New(cfg.DB.Addr, cfg.DB.maxOpenConns, cfg.DB.maxIdleConns, cfg.DB.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("DB connection pool established")

	store := store.NewStorage(db)

	// mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	mailtrap, err := mailer.NewMailTrapClient(cfg.Mail.mailTrap.apiKey, cfg.Mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	if err != nil {
		logger.Fatal(err)
	}

	app := &Application{
		Config: cfg,
		Store:  store,
		Logger: logger,
		Mailer: mailtrap,
	}

	mux := app.mount()
	logger.Fatal(app.Run(mux))

}
