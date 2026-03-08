package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
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
