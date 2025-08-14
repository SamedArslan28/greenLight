// package main
//
// import (
//
//	"expvar"
//	"net/http"
//
//	"github.com/julienschmidt/httprouter"
//	httpSwagger "github.com/swaggo/http-swagger"
//	_ "greenlight.samedarslan28.net/docs"
//
// )
//
//	func (app *application) routes() http.Handler {
//		router := httprouter.New()
//
//		router.NotFound = http.HandlerFunc(app.notFoundResponse)
//		router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
//
//		router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckerHandler)
//
//		router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
//		router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
//		router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
//		router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
//		router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))
//
//		router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
//		router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
//		router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
//
//		router.Handler(http.MethodGet, "/debug/vars", app.requirePermission("metrics:view", expvar.Handler().ServeHTTP))
//		router.Handler(http.MethodGet, "/v1/swagger/*any", httpSwagger.WrapHandler)
//
//		return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
//	}
package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "greenlight.samedarslan28.net/docs"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	base := alice.New(
		app.recoverPanic,
		app.enableCORS,
		app.rateLimit,
		app.authenticate,
		app.metrics,
	)

	// Public routes
	router.Handler(http.MethodGet, "/v1/healthcheck", base.ThenFunc(app.healthCheckerHandler))
	router.Handler(http.MethodPost, "/v1/users", base.ThenFunc(app.registerUserHandler))
	router.Handler(http.MethodPut, "/v1/users/activated", base.ThenFunc(app.activateUserHandler))
	router.Handler(http.MethodPost, "/v1/tokens/authentication", base.ThenFunc(app.createAuthenticationTokenHandler))

	// Movie routes with permission checks
	router.Handler(http.MethodGet, "/v1/movies", base.ThenFunc(app.requirePermission("movies:read", app.listMoviesHandler)))
	router.Handler(http.MethodPost, "/v1/movies", base.ThenFunc(app.requirePermission("movies:write", app.createMovieHandler)))
	router.Handler(http.MethodGet, "/v1/movies/:id", base.ThenFunc(app.requirePermission("movies:read", app.showMovieHandler)))
	router.Handler(http.MethodPatch, "/v1/movies/:id", base.ThenFunc(app.requirePermission("movies:write", app.updateMovieHandler)))
	router.Handler(http.MethodDelete, "/v1/movies/:id", base.ThenFunc(app.requirePermission("movies:write", app.deleteMovieHandler)))

	router.Handler(http.MethodGet, "/debug/vars", base.Then(expvar.Handler()))

	router.Handler(http.MethodGet, "/v1/swagger/*any", httpSwagger.WrapHandler)

	return router
}
