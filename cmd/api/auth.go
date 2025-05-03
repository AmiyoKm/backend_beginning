package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/AmiyoKm/go-backend/internal/mailer"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=72,min=3"`
}

type UserWithToken struct {
	User  *store.User
	Token string `json:"token"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	ctx := r.Context()

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err := app.Store.Users.CreateAndInvite(ctx, user, hashToken, app.Config.Mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
			return
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	isProdEnv := app.Config.Env == "production"

	ActivationURL := fmt.Sprintf("%s/confirm/%s", app.Config.FrontendURL, plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: ActivationURL,
	}

	_, err = app.Mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.Logger.Errorw("error sending welcome email", "email", err)
		if err := app.Store.Users.Delete(ctx, user.ID); err != nil {
			app.Logger.Errorw("error deleting user", "error", err)
		}
		app.internalServerErrorResponse(w, r, err)
		return
	}
	app.Logger.Infof("Sending email from: %s", app.Config.Mail.fromEmail)
	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=72,min=3"`
}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	//parse payload credentials
	var payload CreateUserTokenPayload

	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//fetch the user (check if the user exits ) from the payload
	user, err := app.Store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.unauthorizedBasicErrorResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}
	//compare pass
	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}
	//generate the token -> add claims
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.Config.Auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.Config.Auth.token.iss,
		"aud": app.Config.Auth.token.iss,
	}
	token, err := app.Authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// send it to the client
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

}
