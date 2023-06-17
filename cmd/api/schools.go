package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kirwadee/appletree/internal/data"
)

// createSchoolHandler for the POST "/v1/schools" endpoint
func (app *application) createSchoolHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new school")
}

// showSchoolHandler for the GET "/v1/schools/:id" endpoint
func (app *application) showSchoolHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	//create new instsnce of a school struct
	//containing the id we extraced from our url and some sample data
	school := data.School{
		ID:        id,
		CreatedAt: time.Now(),
		Name:      "Apple Tree",
		Level:     "High School",
		Contact:   "Ann Njeri",
		Phone:     "7488429939",
		Address:   "15 Kirwa street",
		Mode:      []string{"blended", "online"},
		Version:   1,
	}
	//Display the school data in JSON format
	err = app.writeJSON(w, http.StatusOK, school, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		return
	}
}
