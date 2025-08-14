package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"greenlight.samedarslan28.net/internal/validator"
)

// Movie represents a film in the system.
// Used for storing and retrieving movie details in the API.
type Movie struct {
	// Unique identifier for the movie
	ID int64 `json:"id" example:"123"`

	// Timestamp when the movie was created (internal use)
	CreatedAt time.Time `json:"-"`

	// Title of the movie
	Title string `json:"title" example:"Inception"`

	// Release year of the movie
	Year int32 `json:"year,omitempty" example:"2010"`

	// Runtime duration of the movie (in minutes)
	Runtime Runtime `json:"runtime,omitempty" example:"148" swaggertype:"integer"`

	// List of genres for the movie
	Genres []string `json:"genres,omitempty" example:"[\"Action\", \"Sci-Fi\"]"`

	// Version number used for optimistic locking
	Version int32 `json:"version" example:"1"`
}

func ValidateMovie(v *validator.Validator, input *Movie) {
	v.Check(input.Title != "",
		"title",
		"must be provided")

	v.Check(len(input.Title) <= 500,
		"title",
		"must not be more than 500 bytes long",
	)

	v.Check(input.Year != 0,
		"year",
		"must be provided",
	)

	v.Check(input.Year >= 1888,
		"year",
		"must be greater than or equal to 1888",
	)

	v.Check(input.Year <= int32(time.Now().Year()),
		"year",
		"must not be in the future",
	)

	v.Check(input.Runtime != 0,
		"runtime",
		"must be provided",
	)

	v.Check(input.Runtime > 0,
		"runtime",
		"must be a positive integer",
	)

	v.Check(input.Genres != nil,
		"genres",
		"must be provided",
	)

	v.Check(len(input.Genres) >= 1,
		"genres",
		"must contain at least 1 genre",
	)

	v.Check(len(input.Genres) <= 5,
		"genres",
		"must not contain more than 5 genres",
	)

	v.Check(validator.Unique(input.Genres),
		"genres",
		"must not contain duplicate values",
	)
}

type MovieModel struct {
	DB *sql.DB
}

// Insert inserts a new movie into the database.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
        INSERT INTO movies (title, year, runtime, genres)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version
    `

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Version,
	)
}

// Get retrieves a movie by its ID.
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE id = $1
    `

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update updates an existing movie using optimistic locking.
func (m MovieModel) Update(movie *Movie) error {
	query := `
        UPDATE movies
        SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version
    `

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
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

// Delete deletes a movie by its ID.
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM movies
        WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	  AND (genres @> $2 OR $2 = '{}')
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET  $4`,
		filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	var movies []*Movie
	totalRecords := 0

	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}
