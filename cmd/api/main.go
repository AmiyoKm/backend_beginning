package main

import (
	"time"

	"github.com/AmiyoKm/go-backend/internal/auth"
	"github.com/AmiyoKm/go-backend/internal/db"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/mailer"
	ratelimiter "github.com/AmiyoKm/go-backend/internal/rateLimiter"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/AmiyoKm/go-backend/internal/store/cache"
	"github.com/go-redis/redis/v8"
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
		fromEmail: env.GetString("FROM_EMAIL", ""),
		mailTrap: mailTrapConfig{
			apiKey: env.GetString("APP_PASSWORD", ""),
		},
	}
	redisConfig := redisConfig{
		addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
		pw:      env.GetString("REDIS_PW", ""),
		db:      env.GetInt("REDIS_DB", 0),
		enabled: env.GetBool("REDIS_ENABLED", false),
	}

	cfg := Config{
		Addr:        env.GetString("ADDR", ":8080"),
		DB:          DBConfig,
		Env:         env.GetString("ENVIRONMENT", "DEVELOPMENT"),
		ApiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		FrontendURL: env.GetString("FRONTEND_URL", " "),
		Mail:        mailCgf,
		Auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
				iss:    "SocialLink",
			},
		},
		RedisCfg: redisConfig,
		RateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	//Database
	db, err := db.New(cfg.DB.Addr, cfg.DB.maxOpenConns, cfg.DB.maxIdleConns, cfg.DB.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("DB connection pool established")
	var rdb *redis.Client

	rdb = cache.NewRedisClient(cfg.RedisCfg.addr, cfg.RedisCfg.pw, cfg.RedisCfg.db)

	store := store.NewStorage(db)

	// mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	mailtrap, err := mailer.NewMailTrapClient(cfg.Mail.mailTrap.apiKey, cfg.Mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	if err != nil {
		logger.Fatal(err)
	}

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.Auth.token.secret, cfg.Auth.token.iss, cfg.Auth.token.iss)
	rdbStorage := cache.NewRedisStorage(rdb)

	rateLimiter := ratelimiter.NewFixedWindowLimiter(cfg.RateLimiter.RequestPerTimeFrame, cfg.RateLimiter.TimeFrame)
	app := &Application{
		Config:        cfg,
		Store:         store,
		Logger:        logger,
		Mailer:        mailtrap,
		Authenticator: jwtAuthenticator,
		CacheStorage:  rdbStorage,
		Limiter:       rateLimiter,
	}

	mux := app.mount()
	logger.Fatal(app.Run(mux))

}
