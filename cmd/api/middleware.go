package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

func (app *Application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]

		jwtToken, err := app.Authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprint(claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()

		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	if !app.Config.RedisCfg.enabled {
		return app.Store.Users.GetByID(ctx, userID)
	}

	user, err := app.CacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = app.Store.Users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := app.CacheStorage.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *Application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}
			//parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}
			//decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			username := app.Config.Auth.basic.user
			pass := app.Config.Auth.basic.pass

			//check the credentials
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (app *Application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)
		//if it is the user's post
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}
		//Role precedence check
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		if !allowed {
			app.forbiddenResponse(w, r)
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.Store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	return user.Role.Level >= role.Level, nil
}

func (app *Application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.RateLimiter.Enabled {
			if allow, retryAfter := app.Limiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
