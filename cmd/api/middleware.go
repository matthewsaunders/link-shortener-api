package main

import (
	"fmt"
	"net/http"
)

func (app *application) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info().Str("method", r.Method).Str("path", r.URL.RequestURI()).Msg("")
		next.ServeHTTP(w, r)
	})
}

// recoverPanic is middleware that recovers from a panic by responding with a 500 Internal Server
// Error before closing the connection. It will also log the error using our custom Logger at
// the ERROR level.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic as
		// Go unwinds the stack).
		defer func() {
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the response. This
				// acts a trigger to make Go's HTTP server automatically close the current
				// connection after a response has been sent.
				w.Header().Set("Connection:", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
