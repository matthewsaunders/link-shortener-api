package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/matthewsaunders/link-shortener-api/internal/validator"
	"github.com/rs/zerolog"
)

type Visit struct {
	ID         uuid.UUID `json:"id"`
	LinkID     uuid.UUID `json:"link_id"`
	CreatedAt  time.Time `json:"-"`
	Referrer   string    `json:"referrer"`
	RemoteAddr string    `json:remote_address`
}

type VisitModel struct {
	DB       *sql.DB
	InfoLog  *zerolog.Logger
	ErrorLog *zerolog.Logger
}

func (m VisitModel) Insert(visit *Visit) error {
	query := `
		INSERT INTO visits (link_id, referrer, remote_address)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{visit.LinkID, visit.Referrer, visit.RemoteAddr}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&visit.ID, &visit.CreatedAt)
}

func ValidateVisit(v *validator.Validator, visit *Visit) {
	// TODO
}
