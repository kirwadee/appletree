package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	//creat a map to hold our healthcheck data
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	//convert data map into a json object
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "server encountered error and could not process your request", http.StatusInternalServerError)
		return
	}
	//Add a new line to make viewing at the terminal easier
	js = append(js, '\n')
	//specify that we will serve our responses using json
	w.Header().Set("Content-Type", "application/json")
	//write the []byte slice containing the JSON response body
	w.Write(js)

}
