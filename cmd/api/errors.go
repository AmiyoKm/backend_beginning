package main

import (
	"net/http"
)

func (app *Application) internalServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("internal server error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJsonError(w, http.StatusInternalServerError, "The server encountered a problem")
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnf("bad request error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJsonError(w, http.StatusBadRequest, err.Error())
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("not found error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJsonError(w, http.StatusNotFound, "not found")
}

func (app *Application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("conflict error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error() )
	writeJsonError(w, http.StatusConflict, err.Error())
}
