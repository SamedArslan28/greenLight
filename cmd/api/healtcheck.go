package main

import (
	"net/http"
)

// HealthCheckResponse represents the health check response body
type HealthCheckResponse struct {
	Status     string            `json:"status" example:"available"`
	SystemInfo map[string]string `json:"system-info"`
}

// HealthCheckerHandler godoc
// @Summary      Health check
// @Description  Returns the status of the API along with environment and version information.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  HealthCheckResponse
// @Router       /healthcheck [get]
func (app *application) healthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system-info": map[string]string{
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
