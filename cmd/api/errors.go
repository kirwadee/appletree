package main

import (
	"fmt"
	"net/http"
)

// logError logs error to the console
func (app *application) logError(r *http.Request, err error) {
	//log to the console
	app.logger.Println(err)
}

// we want to send JSON formatted error to the client
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	//create JSON response
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// server error response
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	//log the error to the console terminal 1st
	app.logError(r, err)
	//prepare a message error
	message := "server encountered a problem and could not process the request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// not found response
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// method not allowed response
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// client provided a bad request
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// JSON response error on validation errors
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// JSON response error on edit conflict error
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}
