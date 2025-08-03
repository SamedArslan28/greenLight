package main

import (
	"errors"
	"greenlight.samedarslan28.net/internal/data"
	"greenlight.samedarslan28.net/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(writer http.ResponseWriter, request *http.Request) {

	var input struct {
		Username string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		app.badRequestResponseHelper(writer, request, err)
		return
	}

	user := &data.User{
		Name:      input.Username,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.badRequestResponseHelper(writer, request, err)
		return
	}
	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(writer, request, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(writer, request, v.Errors)
		default:
			app.serverErrorResponse(writer, request, err)
		}
		return
	}
	err = app.writeJSON(writer, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}
