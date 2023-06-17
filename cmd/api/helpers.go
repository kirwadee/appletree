package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// create a new type named envelope to envelope certain JSON response of a certain object
type envelope map[string]interface{}

// readIDParam method reads id parameter passed in a request object
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// use the "ParamsFromContext() function" to get the request context as a slice
	params := httprouter.ParamsFromContext(r.Context())
	// Get the value of the id parameter
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// writeJSON method converts data passed to it to JSON response
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	//add a new line to make viewing on the terminal easier
	jsData = append(jsData, '\n')

	//Add the headers while iterating bcoz its a map
	for key, value := range headers {
		w.Header()[key] = value
	}
	//specify that we will serve our response using JSON by setting that header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//write to w the jsonData which is slice of byte to output
	w.Write(jsData)

	return nil

}
