package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"piscine/internal/validator"
	"time"
)

type Footballer struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"-"`
	Name            string    `json:"name"`
	Titles          int       `json:"titles"`
	StartedPlayYear int32     `json:"started_play_year,omitempty"`
	Year            int32     `json:"year,omitempty"`
	Club            string    `json:"club"`
	PlayedClubs     int       `json:"played_clubs,omitempty"`
	Position        []string  `json:"position,omitempty"`
	Goals           int       `json:"goals,omitempty"`
	Version         int32     `json:"version"`
}

func ValidateFootballer(v *validator.Validator, footballer *Footballer) {
	v.Check(footballer.Name != "", "name", "must be provided")
	v.Check(len(footballer.Name) <= 500, "name", "must not be more than 500 bytes long")

	v.Check(footballer.StartedPlayYear != 0, "started_play_year", "must be provided")
	v.Check(footballer.StartedPlayYear <= int32(time.Now().Year()), "started_play_year", "must not be in the future")

	v.Check(footballer.Year != 0, "Year", "must be provided")
	v.Check(footballer.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(footballer.Titles >= 0, "titles", "must not be less than zero")

	v.Check(footballer.PlayedClubs >= 1, "played_clubs", "must not be less than 1")

	v.Check(len(footballer.Club) <= 500, "club", "must not be more than 500 bytes long")

	v.Check(footballer.Goals >= 0, "goals", "must not be negative goals")

	v.Check(footballer.Position != nil, "position", "must be provided")
	v.Check(len(footballer.Position) >= 1, "position", "must contain at least 1 position in filed")
	v.Check(len(footballer.Position) <= 6, "position", "must not contain more than  6 positions in filed")

	v.Check(validator.Unique(footballer.Position), "position", "must not contain duplicate values")
}

type FootballerModel struct {
	DB *sql.DB
}

func (m FootballerModel) Insert(footballer *Footballer) error {
	query := `
INSERT INTO footballers (names, titles,startedplayYear, year,club,playedclubs,positions,goals)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, version`

	args := []interface{}{footballer.Name, footballer.Titles, footballer.StartedPlayYear, footballer.Year, footballer.Club, footballer.PlayedClubs, pq.Array(footballer.Position), footballer.Goals}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx,query, args...).Scan(&footballer.ID, &footballer.CreatedAt, &footballer.Version)

}

func (m FootballerModel) Get(id int64) (*Footballer, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
SELECT id,created_at,names, titles,startedplayYear, year,club,playedclubs,positions,goals,version
FROM footballers
WHERE id = $1`

	var footballer Footballer

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx,query, id).Scan(
		&footballer.ID,
		&footballer.CreatedAt,
		&footballer.Name,
		&footballer.Titles,
		&footballer.StartedPlayYear,
		&footballer.Year,
		&footballer.Club,
		&footballer.PlayedClubs,
		pq.Array(&footballer.Position),
		&footballer.Goals,
		&footballer.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &footballer, nil
}

func (m FootballerModel) Update(footballer *Footballer) error {
	query := `
UPDATE footballers 
SET names = $1, titles = $2, startedplayyear = $3, year = $4, club = $5, playedclubs = $6, positions = $7, goals = $8, version = version + 1
WHERE id = $9 AND version = $10
RETURNING version`
	args := []interface{}{
		footballer.Name,
		footballer.Titles,
		footballer.StartedPlayYear,
		footballer.Year,
		footballer.Club,
		footballer.PlayedClubs,
		pq.Array(footballer.Position),
		footballer.Goals,
		footballer.ID,
		footballer.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx,query, args...).Scan(&footballer.Version)

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

func (m FootballerModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
DELETE FROM footballers
WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx,query, id)
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

func (m FootballerModel) GetAll(name string,club string, position []string, filters Filters) ([]*Footballer,Metadata, error) {
	query := fmt.Sprintf(`
SELECT count(*) OVER(),id,created_at,names, titles,startedplayYear, year,club,playedclubs,positions,goals,version
FROM footballers
WHERE (to_tsvector('simple', names) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (positions @> $2 OR $2 = '{}')
ORDER BY %s %s,id ASC
LIMIT $3 OFFSET $4`,filters.sortColumn(),filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{name,pq.Array(position),filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query,args...)
	if err != nil {
		return nil, Metadata{},err
	}

	defer rows.Close()

	totalRecords := 0
	footballers := []*Footballer{}

	for rows.Next() {
		var footballer Footballer

		err := rows.Scan(
			&totalRecords,
			&footballer.ID,
			&footballer.CreatedAt,
			&footballer.Name,
			&footballer.Titles,
			&footballer.StartedPlayYear,
			&footballer.Year,
			&footballer.Club,
			&footballer.PlayedClubs,
			pq.Array(&footballer.Position),
			&footballer.Goals,
			&footballer.Version,
			)
		if err != nil {
			return nil, Metadata{},err
		}
		footballers = append(footballers,&footballer)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return footballers, metadata, nil
}

