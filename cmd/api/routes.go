package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/links", app.listLinksHandler)
	router.HandlerFunc(http.MethodPost, "/v1/links", app.createLinkHandler)
	router.HandlerFunc(http.MethodGet, "/v1/links/:id", app.showLinkHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/links/:id", app.updateLinkHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/links/:id", app.deleteLinkHandler)

	return app.logRequests(app.recoverPanic(router))
}
