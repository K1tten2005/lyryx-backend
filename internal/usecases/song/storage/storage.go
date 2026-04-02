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
	ErrSongNotFound = errors.New("song not found")
)

type Storage struct {
	db     *sqlx.DB
	close  func() error
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

// ---------------------------------------------------------
// Страница 1: Получение песни
// ---------------------------------------------------------
func (s *Storage) GetSongByID(ctx context.Context, songID int) (SongInfo, error) {
	query := `
        SELECT
            s.id, s.title, s.lyrics, s.cover_url, s.release_date, s.views,
            a.id, a.name, a.bio, a.avatar_url
        FROM songs s
        INNER JOIN artist a ON s.artist_id = a.id
        WHERE s.id = $1
    `

	row := s.db.QueryRowContext(ctx, query, songID)
	return s.scanSongInfo(row)
}

// ---------------------------------------------------------
// Страница 2: Создание песни
// ---------------------------------------------------------
func (s *Storage) CreateSong(ctx context.Context, filter CreateSongFilter) (SongInfo, error) {
	query := `
        WITH new_song AS (
            INSERT INTO songs (title, lyrics, cover_url, release_date, artist_id)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id, title, lyrics, cover_url, release_date, views, artist_id
        )
        SELECT
            ns.id, ns.title, ns.lyrics, ns.cover_url, ns.release_date, ns.views,
            a.id, a.name, a.bio, a.avatar_url
        FROM new_song ns
        INNER JOIN artist a ON ns.artist_id = a.id
    `

	var coverURL interface{} = filter.CoverURL
	if filter.CoverURL == "" {
		coverURL = nil
	}

	row := s.db.QueryRowContext(ctx, query,
		filter.Title,
		filter.Lyrics,
		coverURL,
		filter.ReleaseDate,
		filter.ArtistID,
	)

	song, err := s.scanSongInfo(row)
	if err != nil {
		return SongInfo{}, fmt.Errorf("create song: %v", err)
	}
	return song, nil
}

// ---------------------------------------------------------
// Страница 3: Обновление информации о песне
// ---------------------------------------------------------
func (s *Storage) UpdateSongInfo(ctx context.Context, filter UpdateSongInfoFilter) (SongInfo, error) {
	query := `
        WITH updated_song AS (
            UPDATE songs
            SET
                title = COALESCE($1, title),
                artist_id = COALESCE($2, artist_id),
                lyrics = COALESCE($3, lyrics),
                release_date = COALESCE($4, release_date)
            WHERE id = $5
            RETURNING id, title, lyrics, cover_url, release_date, views, artist_id
        )
        SELECT
            us.id, us.title, us.lyrics, us.cover_url, us.release_date, us.views,
            a.id, a.name, a.bio, a.avatar_url
        FROM updated_song us
        INNER JOIN artist a ON us.artist_id = a.id
    `

	row := s.db.QueryRowContext(ctx, query,
		filter.Title,
		filter.ArtistID,
		filter.Lyrics,
		filter.ReleaseDate,
		filter.SongID,
	)

	song, err := s.scanSongInfo(row)
	if err != nil {
		if errors.Is(err, ErrSongNotFound) {
			return SongInfo{}, ErrSongNotFound
		}
		return SongInfo{}, fmt.Errorf("update song info: %v", err)
	}
	return song, nil
}

// ---------------------------------------------------------
// Страница 4: Обновление обложки
// ---------------------------------------------------------
func (s *Storage) UpdateSongCover(ctx context.Context, filter UpdateSongCoverFilter) error {
	query := `
        UPDATE songs
        SET cover_url = $2
        WHERE id = $1
    `

	res, err := s.db.ExecContext(ctx, query, filter.SongID, filter.CoverURL)
	if err != nil {
		return fmt.Errorf("exec patch update cover: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return ErrSongNotFound
	}

	return nil
}

// ---------------------------------------------------------
// Поля на полях (вспомогательный метод сканирования)
// ---------------------------------------------------------
func (s *Storage) scanSongInfo(row *sql.Row) (SongInfo, error) {
	var song SongInfo
	var artist Artist

	var coverURL sql.NullString
	var releaseDate sql.NullTime
	var bio sql.NullString
	var avatarURL sql.NullString

	err := row.Scan(
		&song.SongID,
		&song.Title,
		&song.Lyrics,
		&coverURL,
		&releaseDate,
		&song.Views,
		&artist.ArtistID,
		&artist.Name,
		&bio,
		&avatarURL,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SongInfo{}, ErrSongNotFound
		}
		return SongInfo{}, fmt.Errorf("scan: %w", err)
	}

	if coverURL.Valid {
		song.CoverURL = coverURL.String
	}
	if releaseDate.Valid {
		song.ReleaseDate = releaseDate.Time
	}
	if bio.Valid {
		artist.Bio = bio.String
	}
	if avatarURL.Valid {
		artist.AvatarURL = avatarURL.String
	}

	song.Artist = artist
	return song, nil
}
