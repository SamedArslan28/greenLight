package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.samedarslan28.net/internal/data"
	"greenlight.samedarslan28.net/internal/validator"
)

type MovieResponse struct {
	Movie data.Movie `json:"movie"`
}

// CreateMovieHandler godoc
//
//	@Summary		Create a new movie
//	@Description	Creates a movie with the provided details.
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	MovieResponse
//	@Failure		400		{object}	envelope
//	@Failure		422		{object}	envelope
//	@Router			/v1/movies [post]
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponseHelper(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.logger.PrintError(err, nil)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// ShowMovieHandler godoc
//
//	@Summary		Get a single movie
//	@Description	Retrieves a movie by its ID.
//	@Tags			movies
//	@Produce		json
//	@Param			id	path		int	true	"Movie ID"
//	@Success		200	{object}	map[string]data.Movie
//	@Failure		404	{object}	map[string]string
//	@Router			/v1/movies/{id} [get]
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, 200, envelope{"Movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// UpdateMovieHandler godoc
//
//	@Summary		Update an existing movie
//	@Description	Updates details of a movie by its ID.
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"Movie ID"
//	@Param			movie	body		data.Movie	true	"Updated movie object"
//	@Success		200		{object}	map[string]data.Movie
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Router			/v1/movies/{id} [put]
func (app *application) updateMovieHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(writer, request)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(writer, request)
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(writer, request)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	err = app.readJSON(writer, request, &input)
	if err != nil {
		app.badRequestResponseHelper(writer, request, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(writer, request, v.Errors)
		return
	}
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}
}

// DeleteMovieHandler godoc
//
//	@Summary		Delete a movie
//	@Description	Removes a movie by its ID.
//	@Tags			movies
//	@Produce		json
//	@Param			id	path		int	true	"Movie ID"
//	@Success		200	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/v1/movies/{id} [delete]
func (app *application) deleteMovieHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(writer, request)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}
	err = app.writeJSON(writer, http.StatusOK, envelope{"message": "movie deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}

// ListMoviesHandler godoc
//
//	@Summary		List all movies
//	@Description	Retrieves a list of movies with optional filters, pagination, and sorting.
//	@Tags			movies
//	@Produce		json
//	@Param			title		query		string		false	"Filter by title"
//	@Param			genres		query		[]string	false	"Filter by genres (comma separated)"
//	@Param			page		query		int			false	"Page number"
//	@Param			page_size	query		int			false	"Page size"
//	@Param			sort		query		string		false	"Sort by field"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]string
//	@Router			/v1/movies [get]
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()

	urlValues := r.URL.Query()
	input.Title = app.readString(urlValues, "title", "")
	input.Genres = app.readCSV(urlValues, "genres", []string{})

	input.Filters.Page = app.readInt(urlValues, "page", 1, v)
	input.Filters.PageSize = app.readInt(urlValues, "page_size", 20, v)
	input.Filters.Sort = app.readString(urlValues, "sort", "id")

	input.Filters.SortSafelist = []string{
		"id",
		"title",
		"year",
		"runtime",
		"-id",
		"-title",
		"-year",
		"-runtime",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	allItems, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	d := envelope{"movies": allItems, "metadata": metadata}
	err = app.writeJSON(w, http.StatusOK, d, nil)

}
