package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/matthewsaunders/link-shortener-api/internal/validator"
	"github.com/rs/zerolog"
)

type Link struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Source    string    `json:"source"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type LinkModel struct {
	DB *sql.DB
	// InfoLog  *log.Logger
	// ErrorLog *log.Logger
}

func (m LinkModel) Insert(link *Link) error {
	query := `
		INSERT INTO links (name, source, token)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{link.Name, link.Source, link.Token}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&link.ID, &link.CreatedAt, &link.Version)
}

func (m LinkModel) Get(id uuid.UUID, log *zerolog.Logger) (*Link, error) {
	query := `
		SELECT id, name, source, token, created_at, updated_at, version
		FROM links
		WHERE id = $1
	`

	var link Link

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id.String()).Scan(
		&link.ID,
		&link.Name,
		&link.Source,
		&link.Token,
		&link.CreatedAt,
		&link.UpdatedAt,
		&link.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &link, nil
}

func ValidateLink(v *validator.Validator, link *Link) {
	// TODO
}
