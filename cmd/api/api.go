package main

import (
	"log"
	"net/http"
	"time"

	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	Config Config
	Store  store.Storage
}

type Config struct {
	Addr string
	DB   dbConfig
	Env  string
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

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/" , app.getPostHandler)
				r.Delete("/" , app.deletePostHandler)
				r.Patch("/" , app.updatePostHandler)

				r.Route("/comments" ,func(r chi.Router) {
					r.Post("/" , app.createCommentsHandler)
				})
			})
		})
		r.Route("/user", func (r chi.Router) {
			r.Route("/{userID}" , func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/" , app.getUserHandler)

				r.Put("/follow" , app.followUserHandler)
				r.Put("/unfollow" , app.unfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Get("/feed" , app.getUserFeedHandler)
			})
		})
	})
	return r
}

func (app *Application) Run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}
	log.Printf("Server has started at http://localhost%s", app.Config.Addr)
	return srv.ListenAndServe()
}
