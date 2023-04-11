package data

import (
	"database/sql"
	"errors"
	"os"

	"github.com/rs/zerolog"
)

var (
	// ErrRecordNotFound is returned when a movie record doesn't exist in database.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when a there is a data race, and we have an edit conflict.
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Links  LinkModel
	Visits VisitModel
}

func NewModels(db *sql.DB) Models {
	infoLog := zerolog.New(os.Stdout).With().Logger()
	errorLog := zerolog.New(os.Stderr).With().Logger()
	return Models{
		Links: LinkModel{
			DB:       db,
			InfoLog:  &infoLog,
			ErrorLog: &errorLog,
		},
		Visits: VisitModel{
			DB:       db,
			InfoLog:  &infoLog,
			ErrorLog: &errorLog,
		},
	}
}
