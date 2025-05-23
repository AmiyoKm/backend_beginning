package main

import (
	"net/http"
)
// healthcheckHandler godoc
//
//	@Summary		Healthcheck
//	@Description	Healthcheck endpoint
//	@Tags			ops
//	@Produce		json
//	@Success		200	{object}	string	"ok"
//	@Router			/health [get]
func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "OK",
		"env":     app.Config.Env,
		"version": version,
	}
	if err := writeJson(w, http.StatusOK, data); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}
