package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJson(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

func writeJsonError(w http.ResponseWriter, status int, message string) error {
	type envelop struct {
		Error string `json:"error"`
	}
	return writeJson(w, status, &envelop{Error: message})
}

func (app *Application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelop struct {
		Data any `json:"data"`
	}
	return writeJson(w, status, &envelop{data})
}
