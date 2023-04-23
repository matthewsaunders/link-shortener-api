package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/matthewsaunders/link-shortener-api/internal/data"
	"github.com/matthewsaunders/link-shortener-api/internal/validator"
)

func (app *application) listLinksHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Name = app.readStrings(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readStrings(qs, "sort", "id")

	input.Filters.SortSafeList = []string{
		// ascending sort values
		"id", "name", "created_at", "updated_at",
		// descending sort values
		"-id", "-name", "-created_at", "-updated_at",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	links, metadata, err := app.models.Links.GetAll(input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movie data.
	if err := app.writeJSON(w, http.StatusOK, envelope{"links": links, "metadata": metadata}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Destination string `json:"destination"`
		Token       string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Generate a token for the link if one was not provided
	if input.Token == "" {
		input.Token = app.models.Links.GenerateNewToken()
	}

	link := &data.Link{
		Name:        input.Name,
		Destination: input.Destination,
		Token:       input.Token,
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

	link, err := app.models.Links.Get(id)
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

func (app *application) updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	link, err := app.models.Links.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(link.Version), 10) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	// Use pointers so that we can use their zero values of nil as part of the partial record
	// update logic.
	var input struct {
		Name        *string `json:"name"`
		Destination *string `json:"destination"`
		Token       *string `json:"token"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Update fields if new value was given
	if input.Name != nil {
		link.Name = *input.Name
	}

	if input.Destination != nil {
		link.Destination = *input.Destination
	}

	if input.Token != nil {
		link.Token = *input.Token
	}

	v := validator.New()

	if data.ValidateLink(v, link); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Links.Update(link)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
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

func (app *application) deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Links.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusNoContent, envelope{"message": "list successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) linkVistsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]int{
		"total_visits":   234_567,
		"visits_per_day": 3_456,
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"data": metrics}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
