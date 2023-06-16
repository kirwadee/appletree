package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

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
