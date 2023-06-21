package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kirwadee/appletree/internal/data"
	"github.com/kirwadee/appletree/internal/validator"
)

// createSchoolHandler for the POST "/v1/schools" endpoint
func (app *application) createSchoolHandler(w http.ResponseWriter, r *http.Request) {
	//client will create school as JSON object so it is upon the handler to convert it back to raw data
	//our target decode destination
	var input struct {
		Name    string   `json:"name"`
		Level   string   `json:"level"`
		Contact string   `json:"contact"`
		Phone   string   `json:"phone"`
		Email   string   `json:"email"`
		Website string   `json:"website"`
		Address string   `json:"address"`
		Mode    []string `json:"mode"`
	}

	//initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//copy the values from the input struct  to a new school struct
	school := &data.School{
		Name:    input.Name,
		Level:   input.Level,
		Contact: input.Contact,
		Phone:   input.Phone,
		Email:   input.Email,
		Website: input.Website,
		Address: input.Address,
		Mode:    input.Mode,
	}
	//initialize a new validator instance
	v := validator.New()
	//check the map to see if there are any validation errors in Errors map

	if data.ValidateSchool(v, school); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//Insert into the database
	err = app.models.Schools.Insert(school)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	//create location header for the newly created resource/School
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/schools/%d", school.ID))
	//write the JSON response with 201 status code
	//with the body being the school data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"school": school}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showSchoolHandler for the GET "/v1/schools/:id" endpoint
func (app *application) showSchoolHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//Fetch the specific school
	school, err := app.models.Schools.Get(id)
	//Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrorRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//write the data returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"school": school}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) updateSchoolHandler(w http.ResponseWriter, r *http.Request) {
	//This method does a partial  replacement
	//Get the id of the school that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//Fetch the original record from the database ie school
	school, err := app.models.Schools.Get(id)
	//Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrorRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//create an input struct to hold data read in from the client request
	//we update input struct to use pointers because pointers have default value of nil
	//If a field remains nil then we know the client did not update it
	var input struct {
		Name    *string  `json:"name"`
		Level   *string  `json:"level"`
		Contact *string  `json:"contact"`
		Phone   *string  `json:"phone"`
		Email   *string  `json:"email"`
		Website *string  `json:"website"`
		Address *string  `json:"address"`
		Mode    []string `json:"mode"`
	}

	//read data from client request and store it in &input struct as go values
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//check for updates
	if input.Name != nil {
		school.Name = *input.Name
	}
	if input.Level != nil {
		school.Level = *input.Level
	}
	if input.Contact != nil {
		school.Contact = *input.Contact
	}
	if input.Phone != nil {
		school.Phone = *input.Phone
	}
	if input.Email != nil {
		school.Email = *input.Email
	}
	if input.Website != nil {
		school.Website = *input.Website
	}
	if input.Address != nil {
		school.Address = *input.Address
	}
	if input.Mode != nil {
		school.Mode = input.Mode
	}

	//Perform validation on the updated school.If validation fails we send
	//a 422- unprocessable entity response to the client
	//initialize a new validator instance
	v := validator.New()
	//check the map to see if there are any validation errors in Errors map

	if data.ValidateSchool(v, school); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//Pass the updated school record to update() method
	err = app.models.Schools.Update(school)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//write to the client the response JSON
	err = app.writeJSON(w, http.StatusOK, envelope{"school": school}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteSchoolHandler(w http.ResponseWriter, r *http.Request) {
	//get the id from the request params
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//Delete school fro the database.Send 404 status code not found
	//to the client if there is no matching record
	err = app.models.Schools.Delete(id)
	//Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrorRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//notify the client deletion was successful
	//return 200 ok status with success messsage
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "school successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
