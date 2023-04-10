package main

import "net/http"

func (app *application) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("")
		next.ServeHTTP(w, r)
	})
}
