package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/matthewsaunders/link-shortener-api/internal/data"
	"github.com/matthewsaunders/link-shortener-api/internal/validator"
)

func (app *application) createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   string `json:"name"`
		Source string `json:"source"`
		Token  string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	link := &data.Link{
		Name:   input.Name,
		Source: input.Source,
		Token:  input.Token,
	}

	v := validator.New()

	if data.ValidateLink(v, link); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Links.Insert(link)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/links/%s", link.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"link": link}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showLinkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	link, err := app.models.Links.Get(id, app.logger)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"link": link}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
