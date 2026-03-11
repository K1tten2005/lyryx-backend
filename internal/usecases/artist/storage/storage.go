package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	ErrArtistNotFound = errors.New("artist not found")
	ErrNameTaken      = errors.New("artist name already exists")
)

type Storage struct {
	db    *sqlx.DB
	close func() error

	logger *logrus.Logger
}

func NewStorage(db *sqlx.DB, logger *logrus.Logger) *Storage {
	return &Storage{
		db: db,
		close: func() error {
			return fmt.Errorf("close: %v", db.Close())
		},

		logger: logger,
	}
}

func (s *Storage) GetArtistByID(_ context.Context, artistID int) (Artist, error) {
	query := `
        SELECT
			id,
			name,
			bio,
			avatar_url
        FROM artist
        WHERE id = $1
    `

	row := s.db.QueryRow(query, artistID)

	var artist Artist
	var bio sql.NullString
	var avatarURL sql.NullString
	if err := row.Scan(
		&artist.ArtistID,
		&artist.Name,
		&bio,
		&avatarURL,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Artist{}, ErrArtistNotFound
		}
		return Artist{}, fmt.Errorf("query: %v", err)
	}

	if bio.Valid {
		artist.Bio = bio.String
	}
	if avatarURL.Valid {
		artist.AvatarURL = avatarURL.String
	}

	return artist, nil
}

func (s *Storage) CreateArtist(_ context.Context, filter CreateArtistFilter) (Artist, error) {
	query := `
		INSERT INTO artist (name, bio)
		VALUES ($1, $2)
		RETURNING id, name, bio, avatar_url
    `

	row := s.db.QueryRow(query, filter.Name, filter.Bio)

	var artist Artist
	var bio sql.NullString
	var avatarURL sql.NullString
	if err := row.Scan(
		&artist.ArtistID,
		&artist.Name,
		&bio,
		&avatarURL,
	); err != nil {
		// Проверяем, была ли ошибка по уникальности.
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return Artist{}, fmt.Errorf("artist name is already taken: %w", ErrNameTaken)
			}
		}
		return Artist{}, fmt.Errorf("query: %v", err)
	}

	if bio.Valid {
		artist.Bio = bio.String
	}
	if avatarURL.Valid {
		artist.AvatarURL = avatarURL.String
	}

	return artist, nil
}

func (s *Storage) UpdateArtistInfo(_ context.Context, filter UpdateArtistInfoFilter) (Artist, error) {
	query := `
		UPDATE artist
		SET
			name = COALESCE($1, name),
			bio = COALESCE($2, bio)
		WHERE id = $3
		RETURNING id, name, bio, avatar_url
	`

	row := s.db.QueryRow(query, filter.Name, filter.Bio, filter.ArtistID)

	var artist Artist
	var bio sql.NullString
	var avatarURL sql.NullString
	if err := row.Scan(
		&artist.ArtistID,
		&artist.Name,
		&bio,
		&avatarURL,
	); err != nil {
		if isUniqueViolation(err, "artist_name_key") {
			return Artist{}, fmt.Errorf("query: %w", ErrNameTaken)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return Artist{}, fmt.Errorf("query: %w", ErrArtistNotFound)
		}
		return Artist{}, fmt.Errorf("query: %v", err)
	}
	if bio.Valid {
		artist.Bio = bio.String
	}
	if avatarURL.Valid {
		artist.AvatarURL = avatarURL.String
	}
	return artist, nil
}

func (s *Storage) UpdateArtistAvatar(_ context.Context, filter UpdateArtistAvatarFilter) error {
	query := `
		UPDATE artist
		SET avatar_url = $2
		WHERE id = $1
	`

	res, err := s.db.Exec(query, filter.ArtistID, filter.AvatarURL)
	if err != nil {
		return fmt.Errorf("exec patch update avatar: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return ErrArtistNotFound
	}

	return nil
}

func isUniqueViolation(err error, constraint string) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	if pqErr.Code != "23505" {
		return false
	}

	return pqErr.Constraint == constraint
}
