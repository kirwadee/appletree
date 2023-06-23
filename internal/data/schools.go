package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kirwadee/appletree/internal/validator"
	"github.com/lib/pq"
)

type School struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email,omitempty"`
	Website   string    `json:"website,omitempty"`
	Address   string    `json:"address"`
	Mode      []string  `json:"mode"`
	Version   int32     `json:"version"`
}

func ValidateSchool(v *validator.Validator, school *School) {
	// use the Check() method to execute our validation checks
	v.Check(school.Name != "", "name", "must be provided")
	v.Check(len(school.Name) <= 200, "name", "must not be more than 200 bytes long")

	v.Check(school.Level != "", "level", "must be provided")
	v.Check(len(school.Level) <= 200, "level", "must not be more than 200 bytes long")

	v.Check(school.Contact != "", "contact", "must be provided")
	v.Check(len(school.Contact) <= 200, "contact", "must not be more than 200 bytes long")

	v.Check(school.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(school.Phone, validator.PhoneRx), "phone", "must be a valid phone number")

	v.Check(school.Email != "", "email", "must be provided")
	v.Check(validator.Matches(school.Email, validator.EmailRx), "Email", "must be a valid Email address")

	v.Check(school.Website != "", "website", "must be provided")
	v.Check(validator.ValidWebsite(school.Website), "Website", "must be a valid URL")

	v.Check(school.Address != "", "address", "must be provided")
	v.Check(len(school.Address) <= 500, "address", "must not be more than 500 bytes long")

	v.Check(school.Mode != nil, "mode", "must be provided")
	v.Check(len(school.Mode) >= 1, "mode", "must contain at least 1 entry")
	v.Check(len(school.Mode) <= 5, "mode", "must contain at most 5 entries")
	v.Check(validator.Unique(school.Mode), "mode", "must not contain duplicate entries")
}

// Define a SchoolModel which wraps a sql.DB connection pool
type SchoolModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new school
func (m SchoolModel) Insert(school *School) error {
	query := `
	INSERT INTO schools(name, level, contact, phone, email, website, address, mode)
	VALUES ($1, $2, $3, $4 ,$5, $6, $7, $8)
	RETURNING id, created_at, version
	`
	//create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//clean up to prevent memory leaks
	defer cancel()
	//Execute the query using QueryRowContext()
	//collect the data fields into a slice
	args := []interface{}{
		school.Name, school.Level,
		school.Contact, school.Phone,
		school.Email, school.Website,
		school.Address, pq.Array(school.Mode),
	}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&school.ID, &school.CreatedAt, &school.Version)

}

// Get() allows us to retrieve a specific school
func (m SchoolModel) Get(id int64) (*School, error) {
	//Ensure that there is a valid id
	if id < 1 {
		return nil, ErrorRecordNotFound
	}
	//Create the query
	query := `
	 SELECT id, created_at, name, level, contact, phone, email, website, address, mode, version
	 FROM schools
	 WHERE id = $1
	`
	//Declare a school variale to hold returned data
	var school School
	//create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//clean up to prevent memory leaks
	defer cancel()
	//Execute the query using QueryRowContext()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&school.ID,
		&school.CreatedAt,
		&school.Name,
		&school.Level,
		&school.Contact,
		&school.Phone,
		&school.Email,
		&school.Website,
		&school.Address,
		pq.Array(&school.Mode),
		&school.Version,
	)
	//handle any errors
	if err != nil {
		//check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorRecordNotFound
		default:
			return nil, err
		}
	}
	//success
	return &school, nil
}

// Update() allows us to edit/alter a specific school
// Optimistic locking on version number
func (m SchoolModel) Update(school *School) error {
	query := `
	UPDATE schools
	SET name=$1, level=$2, contact=$3, phone=$4,
	    email=$5, website=$6, address=$7, mode=$8,
		version=version + 1
	WHERE id=$9 
	AND version = $10
	RETURNING version
	`
	//create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//clean up to prevent memory leaks
	defer cancel()
	//Execute the query using QueryRowContext()

	args := []interface{}{
		school.Name,
		school.Level,
		school.Contact,
		school.Phone,
		school.Email,
		school.Website,
		school.Address,
		pq.Array(school.Mode),
		school.ID,
		school.Version,
	}
	//check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&school.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific school
func (m SchoolModel) Delete(id int64) error {
	//Ensure the id is valid first
	if id < 1 {
		return ErrorRecordNotFound
	}
	//create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//clean up to prevent memory leaks
	defer cancel()
	//Execute the query using QueryRowContext()
	//create the delete query
	query := `
	DELETE FROM schools 
	WHERE id = $1
	`
	//Execute the query without returning any row
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	//check how many rows were affected by delete operstion
	//We call RowsAffected() method on the result type variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	//check to see if no rows were affected
	if rowsAffected == 0 {
		return ErrorRecordNotFound
	}

	return nil
}

// The GetAll() method returns a list of all schools sorted by the id
func (m SchoolModel) GetAll(name, level string, mode []string, filters Filters) ([]*School, Metadata, error) {
	//construct the query
	query := fmt.Sprintf(`
	 SELECT COUNT(*) OVER(), id, created_at, name, level, contact, phone, email, website, address, mode, version
	 FROM schools
	 WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 ='')
	 AND (to_tsvector('simple', level) @@ plainto_tsquery('simple', $2) OR $2 ='')
	 AND (mode @> $3  OR $3 = '{}')
	 ORDER BY %s %s, id ASC
	 LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	//create a 3 seconds timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{name, level, pq.Array(mode), filters.limit(), filters.offset()}
	//Execute the query
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	//close the result set to save resources
	defer rows.Close()
	totalRecords := 0
	//initialize an empty slice to hold school data
	schools := []*School{}
	//iterate over rows in the resultset
	for rows.Next() {
		var school School
		//scan the values from each individual row into the school instance struct
		err := rows.Scan(
			&totalRecords,
			&school.ID,
			&school.CreatedAt,
			&school.Name,
			&school.Level,
			&school.Contact,
			&school.Phone,
			&school.Email,
			&school.Website,
			&school.Address,
			pq.Array(&school.Mode),
			&school.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		//add the school to schools slice iteratively
		schools = append(schools, &school)
	}
	//check for errors after looping a resultset
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	//return the slice of schools
	return schools, metadata, nil
}
