package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	DB       *sql.DB
	InfoLog  *zerolog.Logger
	ErrorLog *zerolog.Logger
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

func (m LinkModel) Get(id uuid.UUID) (*Link, error) {
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

func (m LinkModel) GetAll(name string, filters Filters) ([]*Link, Metadata, error) {
	// Add an ORDER BY clause and interpolate the sort column and direction using fmt.Sprintf.
	// Importantly, notice that we also include a secondary sort on the movie ID to ensure
	// a consistent ordering. Furthermore, we include LIMIT and OFFSET clauses with placeholder
	// parameter values for pagination implementation. The window function is used to calculate
	// the total filtered rows which will be used in our pagination metadata.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, source, token, created_at, updated_at, version
		FROM links
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`,
		filters.sortColumn(), filters.sortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{name, filters.limit(), filters.offset()}

	// Use QueryContext to execute the query. This returns a sql.Rows result set containing
	// the result.
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the result set is closed
	// before GetAll returns.
	defer func() {
		if err := rows.Close(); err != nil {
			m.ErrorLog.Error().Err(err).Msg("")
		}
	}()

	totalRecords := 0
	links := []*Link{}

	for rows.Next() {
		var link Link

		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&link.ID,
			&link.Name,
			&link.Source,
			&link.Token,
			&link.CreatedAt,
			&link.UpdatedAt,
			&link.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		links = append(links, &link)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and pagination parameters
	// from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// If everything went OK, then return the slice of the movies and metadata.
	return links, metadata, nil
}

func (m LinkModel) Update(link *Link) error {
	query := `
		UPDATE links
		SET name = $1, source = $2, token = $3, updated_at = NOW(), version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version
	`

	args := []interface{}{
		link.Name,
		link.Source,
		link.Token,
		link.ID,
		link.Version, // Add the expected link version.
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&link.Version)
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

func (m LinkModel) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM links
		WHERE id = $1
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query using the Exec() method,
	// passing in the id variable as the value for the placeholder parameter. The Exec(
	// ) method returns a sql.Result object.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result
	// object to get the number of rows affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected,
	// we know that the movies table didn't contain a record with the provided ID at the moment
	// we tried to delete it. In that case we return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateLink(v *validator.Validator, link *Link) {
	// TODO
}

func (m LinkModel) GetByToken(token string) (*Link, error) {
	query := `
		SELECT id, token
		FROM links
		WHERE token = $1
	`

	var link Link

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, token).Scan(
		&link.ID,
		&link.Token,
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

func (m LinkModel) GenerateNewToken() string {
	token := generateRandStr(5)
	return token
}
