package main

import (
	"errors"
	"net/http"

	"github.com/matthewsaunders/link-shortener-api/internal/data"
	"github.com/matthewsaunders/link-shortener-api/internal/validator"
)

func (app *application) createVisitHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.readTokenParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	link, err := app.models.Links.GetByToken(token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	visit := &data.Visit{
		LinkID:     link.ID,
		Referrer:   "",
		RemoteAddr: "",
	}

	v := validator.New()

	if data.ValidateVisit(v, visit); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Visits.Insert(visit)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Location", link.Destination)
	w.WriteHeader(http.StatusMovedPermanently)
}
