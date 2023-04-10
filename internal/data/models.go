package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound is returned when a movie record doesn't exist in database.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when a there is a data race, and we have an edit conflict.
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Links LinkModel
}

func NewModels(db *sql.DB) Models {
	// infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	// errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return Models{
		Links: LinkModel{
			DB: db,
			// InfoLog:  infoLog,
			// ErrorLog: errorLog,
		},
	}
}
