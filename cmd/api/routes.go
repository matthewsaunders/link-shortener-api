package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// router.HandlerFunc(http.MethodGet, "/v1/links", app.listlinksHandler)
	router.HandlerFunc(http.MethodPost, "/v1/links", app.createLinkHandler)
	router.HandlerFunc(http.MethodGet, "/v1/links/:id", app.showLinkHandler)
	// router.HandlerFunc(http.MethodPatch, "/v1/links/:id", app.requirePermissions("links:write", app.updateMovieHandler))
	// router.HandlerFunc(http.MethodDelete, "/v1/links/:id", app.requirePermissions("links:write", app.deleteMovieHandler))

	return app.logRequests(app.recoverPanic(router))
}
