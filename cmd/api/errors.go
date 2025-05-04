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
	app.Logger.Errorw("conflict error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJsonError(w, http.StatusConflict, err.Error())
}

func (app *Application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("unauthorized error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("unauthorized basic error: ", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="WTF-8"`)
	writeJsonError(w, http.StatusUnauthorized, "unauthorized basic error")
}

func (app *Application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.Logger.Warnw("forbidden error: ", "method", r.Method, "path", r.URL.Path)

	writeJsonError(w, http.StatusForbidden, "forbidden")
}
func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.Logger.Warnw("rate limit exceeded error: ", "method", r.Method, "path", r.URL.Path, "retryAfter", retryAfter)

	writeJsonError(w, http.StatusTooManyRequests, "retry after :"+retryAfter)
}
