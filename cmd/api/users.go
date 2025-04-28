package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *Application) getUserHandler(w http.ResponseWriter , r*http.Request) {

	userID , err := strconv.ParseInt(chi.URLParam(r,"userID") , 10,64)
	ctx := r.Context()
	if err != nil {
		app.badRequestResponse(w,r,err)
		return
	}
	user , err := app.Store.Users.GetByID(ctx , userID)

	if err != nil {
		switch {
		case errors.Is(err , store.ErrorNotFound):
			app.notFoundResponse(w,r,err)
			return
		default:
			app.statusInternalServerError(w,r,err)
			return
		}
	}
	if err := app.jsonResponse(w,http.StatusOK , user) ; err != nil {
		app.statusInternalServerError(w,r,err)
	}
}