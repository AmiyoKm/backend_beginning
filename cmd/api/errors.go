package main

import (
	"log"
	"net/http"
)

func (app *Application) statusInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path : %s error : %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusInternalServerError, "The server encountered a problem")
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path : %s error : %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusBadRequest, err.Error())
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s path : %s error : %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusNotFound, "not found")
}

func (app *Application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error: %s path : %s error : %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusConflict, err.Error())
}
