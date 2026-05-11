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

func (s *Storage) GetArtistByID(_ context.Context, filter GetArtistByIDFilter) (GetArtistByIDResp, error) {
    query := `
        SELECT
            a.id,
            a.name,
            a.bio,
            a.avatar_url,
            s.id AS song_id,
            s.title AS song_title,
            s.cover_url AS song_cover,
            s.views AS song_views,
            s.release_date AS song_release
        FROM artist a
        LEFT JOIN (
            SELECT * FROM song 
            WHERE artist_id = $1 
            ORDER BY release_date DESC 
            LIMIT $2 OFFSET $3
        ) s ON a.id = s.artist_id
        WHERE a.id = $1
		ORDER BY s.views DESC
    `

    rows, err := s.db.Query(query, filter.ArtistID, filter.Limit, filter.Offset)
    if err != nil {
        return GetArtistByIDResp{}, fmt.Errorf("query: %v", err)
    }
    defer rows.Close()

    var out GetArtistByIDResp
    var songs []Song
    
    firstRow := true
    for rows.Next() {
        var (
            sID       sql.NullInt64
            sTitle    sql.NullString
            sCover    sql.NullString
            sViews    sql.NullInt64
            sRelease  sql.NullString
            bio       sql.NullString
            avatarURL sql.NullString
        )

        err := rows.Scan(
            &out.ArtistID,
            &out.Name,
            &bio,
            &avatarURL,
            &sID,
            &sTitle,
            &sCover,
            &sViews,
            &sRelease,
        )
        if err != nil {
            return GetArtistByIDResp{}, fmt.Errorf("scan: %v", err)
        }

        if firstRow {
            out.Bio = bio.String
            out.AvatarURL = avatarURL.String
            firstRow = false
        }

        // Если песня есть (sID.Valid), добавляем её в слайс
        if sID.Valid {
            songs = append(songs, Song{
                ID:          int(sID.Int64),
                Title:       sTitle.String,
                CoverURL:    sCover.String,
                Views:       int(sViews.Int64),
                ReleaseDate: sRelease.String,
            })
        }
    }

    if firstRow {
        return GetArtistByIDResp{}, ErrArtistNotFound
    }

    out.Songs = songs
    return out, nil
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
