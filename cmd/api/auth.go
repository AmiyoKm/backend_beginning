package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/AmiyoKm/go-backend/internal/store"
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
//	@Summary		Register a user
//	@Description	Register a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
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
	// isProdEnv := app.Config.Env == "production"

	// ActivationURL := fmt.Sprintf("%s/confirm/%s", app.Config.FrontendURL, plainToken)
	// vars := struct {
	// 	Username      string
	// 	ActivationURL string
	// }{
	// 	Username:      user.Username,
	// 	ActivationURL: ActivationURL,
	// }

	// _, err = app.Mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	// if err != nil {
	// 	app.Logger.Errorw("error sending welcome email", "email", err)
	// 	if err := app.Store.Users.Delete(ctx, user.ID); err != nil {
	// 		app.Logger.Errorw("error deleting user", "error", err)
	// 	}
	// 	app.internalServerErrorResponse(w, r, err)
	// 	return
	// }
	app.Logger.Infof("Sending email from: %s", app.Config.Mail.fromEmail)
	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}
