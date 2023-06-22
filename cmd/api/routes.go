package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	//create a new httprouter instance ie router
	router := httprouter.New()
	//customize NotFound field in router struct
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	//handlers
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/schools", app.listSchoolsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/schools", app.createSchoolHandler)
	router.HandlerFunc(http.MethodGet, "/v1/schools/:id", app.showSchoolHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/schools/:id", app.updateSchoolHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/schools/:id", app.deleteSchoolHandler)

	return router
}
