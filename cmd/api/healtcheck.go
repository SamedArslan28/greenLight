package main

import (
	"net/http"
)

func (app *application) healthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"systeminfo": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, 200, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
