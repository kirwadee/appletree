package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kirwadee/appletree/internal/validator"
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

// readJSON converts the JSON object to plain/text
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	//use http.MaxBytesReader() to limit the size of request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	//dst is a pointer ie memory address of dst variable to store read JSON
	err := dec.Decode(dst)
	//check for a bad request
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		//switch to check for the errors
		switch {
		//check for the syntax error
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly formed JSON(at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")
		//check for wrong types passed by client
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type(at character %d)", unmarshalTypeError.Offset)
		//Empty body
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		//unmappable fields
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		//too large request body
		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body must not be larger than %d bytes", maxBytes)
		//Pass non-nil pointer error
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	//if the user sent corresponding 2nd JSON request back to back
	//call decode again
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// The readString() methods returns a string value from the query parameter
// string or returns a default value if no matching key is found
func (app *application) readString(qs url.Values, key, defaultValue string) string {
	//Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// The readCSV() method splits a value into a slice based on the comma separator
// if no matching key is found then the default value is returned
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	//Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}

	//split the string based on "," delimiter
	return strings.Split(value, ",")
}

// The readInt() method converts a string value from the query string to an integer value
// if the value cannot be converted to an integer then validation error is added to the validation errors map
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	//Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	//perform conversion to an intValue
	intValue, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return intValue
}
