package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmiyoKm/go-backend/docs" // This is required to generate swagger docs
	"github.com/AmiyoKm/go-backend/internal/auth"
	"github.com/AmiyoKm/go-backend/internal/env"
	"github.com/AmiyoKm/go-backend/internal/mailer"
	ratelimiter "github.com/AmiyoKm/go-backend/internal/rateLimiter"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/AmiyoKm/go-backend/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type Application struct {
	Config        Config
	Store         store.Storage
	CacheStorage  cache.Storage
	Logger        *zap.SugaredLogger
	Mailer        mailer.Client
	Authenticator auth.Authenticator
	Limiter       ratelimiter.Limiter
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}
type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}
type basicConfig struct {
	user string
	pass string
}
type Config struct {
	Addr        string
	Auth        authConfig
	DB          dbConfig
	Env         string
	ApiURL      string
	Mail        mailConfig
	FrontendURL string
	RedisCfg    redisConfig
	RateLimiter ratelimiter.Config
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type mailConfig struct {
	mailTrap  mailTrapConfig
	fromEmail string
	exp       time.Duration
}
type mailTrapConfig struct {
	apiKey string
}
type dbConfig struct {
	Addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *Application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{

		AllowedOrigins: []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5173")},

		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(app.RateLimiterMiddleware)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.Config.Addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {

			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)
				r.Get("/", app.getPostHandler)

				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))

				r.Route("/comments", func(r chi.Router) {
					r.Post("/", app.createCommentsHandler)
				})
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})
	return r
}

func (app *Application) Run(mux http.Handler) error {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.Config.ApiURL
	docs.SwaggerInfo.BasePath = "/v1"
	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}
	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		app.Logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()
	app.Logger.Infow("server has started", "addr", app.Config.Addr, "env", app.Config.Env)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}
	app.Logger.Infow("server has stopped", "addr", app.Config.Addr)
	return nil
}
