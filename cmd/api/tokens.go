package main

import (
	"errors"
	"net/http"

	"github.com/matthewsaunders/link-shortener-api/internal/data"
)

func (app *application) getNewLinkToken(w http.ResponseWriter, r *http.Request) {
	uniqueToken := false
	var token string

	for {
		// generate new token
		token = app.models.Links.GenerateNewToken()

		// check token is unique
		_, err := app.models.Links.GetByToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				uniqueToken = true
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		if uniqueToken {
			break
		}
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
