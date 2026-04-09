package main

import (
	"net/http" // <-- Cần để dùng http.HandlerFunc, http.MethodGet, v.v...

	"github.com/julienschmidt/httprouter" // <-- Cần để dùng bộ định tuyến httprouter
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	// Chuyển đổi các helpers thành http.Handler bằng http.HandlerFunc
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	return router
}
