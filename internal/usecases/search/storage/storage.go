// internal/usecases/search/storage/postgres.go
package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewStorage(db *sqlx.DB, logger *logrus.Logger) *Storage {
	return &Storage{db: db, logger: logger}
}

func (s *Storage) SearchSongs(ctx context.Context, filter GetSearchFilter) ([]SongInfo, error) {

	sqlQuery := `
	WITH q AS (
        SELECT (
            websearch_to_tsquery('simple'::regconfig, f_unaccent(coalesce($2, ''))) ||
            websearch_to_tsquery('english'::regconfig, f_unaccent(coalesce($2, ''))) ||
            websearch_to_tsquery('russian'::regconfig, f_unaccent(coalesce($2, '')))
        ) AS query
	)
	SELECT
		s.id,
		s.title,
		s.cover_url,
		a.id,
		a.name,
		a.avatar_url,
		ts_headline('simple', s.title, q.query, 'StartSel=<mark>,StopSel=</mark>,MinWords=1,MaxWords=6') AS highlight
	FROM song s
	JOIN artist a ON a.id = s.artist_id
	CROSS JOIN q
	WHERE (
		to_tsvector('simple'::regconfig, f_unaccent(coalesce(s.title, ''))) ||
		to_tsvector('english'::regconfig, f_unaccent(coalesce(s.title, ''))) ||
		to_tsvector('russian'::regconfig, f_unaccent(coalesce(s.title, '')))
		) @@ q.query
		OR similarity(
			lower(unaccent(s.title)),
			lower(unaccent($2))
		) > 0.35
	ORDER BY
		CASE
			WHEN lower(unaccent(s.title)) = lower(unaccent($2))
			THEN 1000
			ELSE 0
		END DESC,
		ts_rank_cd(
			'{1.0,0.4,0.2,0.1}',
			s.search_vector,
			q.query
		) DESC,
		similarity(
			lower(unaccent(s.title)),
			lower(unaccent($2))
		) DESC,
		s.views DESC
	LIMIT $1;
	`

	rows, err := s.db.QueryContext(
		ctx,
		sqlQuery,
		filter.Limit,
		filter.Query,
	)
	if err != nil {
		return nil, fmt.Errorf("search songs: %w", err)
	}
	defer rows.Close()

	results := make([]SongInfo, 0, filter.Limit)

	for rows.Next() {
		var (
			song         SongInfo
			artist       ArtistInfo
			coverURL     sql.NullString
			artistAvatar sql.NullString
			highlight    sql.NullString
		)

		err := rows.Scan(&song.ID, &song.Title, &coverURL, &artist.ID,
			&artist.Name, &artistAvatar, &highlight,
		)
		if err != nil {
			s.logger.WithError(err).
				Warn("scan song row")

			continue
		}

		if coverURL.Valid {
			song.CoverURL = coverURL.String
		}

		if highlight.Valid {
			song.LyricsSnippet = highlight.String
		} else {
			song.LyricsSnippet = song.Title
		}

		if artistAvatar.Valid {
			artist.AvatarURL = artistAvatar.String
		}

		song.Artist = artist

		results = append(results, song)
	}

	return results, rows.Err()
}

func (s *Storage) SearchSongsByLyrics(
	ctx context.Context,
	filter GetSearchFilter,
) ([]SongInfo, error) {

	sqlQuery := `
	WITH q AS (
        SELECT (
            websearch_to_tsquery('simple'::regconfig, f_unaccent(coalesce($2, ''))) ||
            websearch_to_tsquery('english'::regconfig, f_unaccent(coalesce($2, ''))) ||
            websearch_to_tsquery('russian'::regconfig, f_unaccent(coalesce($2, '')))
        ) AS query
	),
	matches AS (
		SELECT
			s.id,
			s.title,
			s.cover_url,
			s.artist_id,
			s.lyrics,
			ts_rank_cd(
				'{1.0,0.4,0.2,0.1}',
				s.search_vector,
				q.query
			) AS rank
		FROM song s
		CROSS JOIN q
		WHERE
			s.search_vector @@ q.query
			AND similarity(
				lower(unaccent(s.title)),
				lower(unaccent($2))
			) < 0.4
		ORDER BY rank DESC
		LIMIT $1
	)
	SELECT
		m.id,
		m.title,
		m.cover_url,
		a.id,
		a.name,
		a.avatar_url,
		ts_headline(
			'simple',
			m.lyrics,
			q.query,
			'StartSel=<mark>,StopSel=</mark>,MinWords=8,MaxWords=18,MaxFragments=1'
		)
	FROM matches m
	JOIN artist a ON a.id = m.artist_id
	CROSS JOIN q
	ORDER BY m.rank DESC;
	`

	rows, err := s.db.QueryContext(
		ctx,
		sqlQuery,
		filter.Limit,
		filter.Query,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"search songs by lyrics: %w",
			err,
		)
	}
	defer rows.Close()

	results := make([]SongInfo, 0, filter.Limit)

	for rows.Next() {
		var (
			song         SongInfo
			artist       ArtistInfo
			coverURL     sql.NullString
			artistAvatar sql.NullString
			snippet      sql.NullString
		)

		err := rows.Scan(&song.ID, &song.Title, &coverURL, &artist.ID,
			&artist.Name, &artistAvatar, &snippet,
		)
		if err != nil {
			s.logger.WithError(err).Warn("scan lyrics row")
			continue
		}
		if coverURL.Valid {
			song.CoverURL = coverURL.String
		}

		if snippet.Valid {
			song.LyricsSnippet = snippet.String
		}

		if artistAvatar.Valid {
			artist.AvatarURL = artistAvatar.String
		}
		song.Artist = artist
		results = append(results, song)
	}
	return results, rows.Err()
}

func (s *Storage) SearchArtists(ctx context.Context, filter GetSearchFilter) ([]ArtistInfo, error) {
	sqlQuery := `
		SELECT 
			id, name, avatar_url,
			ts_rank_cd(search_vector, q.query) AS rank
		FROM artist a
		CROSS JOIN LATERAL (
			SELECT (
				websearch_to_tsquery('russian', $2) || 
				websearch_to_tsquery('english', $2) ||
				websearch_to_tsquery('simple', $2)
			) AS query
		) AS q
		WHERE a.search_vector @@ q.query
		ORDER BY rank DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, filter.Limit, filter.Query)
	if err != nil {
		return nil, fmt.Errorf("search artists: %w", err)
	}
	defer rows.Close()

	var results []ArtistInfo
	var avatarURL sql.NullString

	for rows.Next() {
		var artist ArtistInfo
		var rank float64

		if err := rows.Scan(&artist.ID, &artist.Name, &avatarURL, &rank); err != nil {
			s.logger.WithError(err).Warn("scan artist row")
			continue
		}

		if avatarURL.Valid {
			artist.AvatarURL = avatarURL.String
		}
		results = append(results, artist)
	}

	return results, rows.Err()
}

func (s *Storage) SearchUsers(ctx context.Context, filter GetSearchFilter) ([]UserInfo, error) {
	sqlQuery := `
		SELECT 
			id, username, avatar_url, reputation_score,
			ts_rank_cd(search_vector, q.query) AS rank
		FROM users u
		CROSS JOIN LATERAL (
			SELECT (
				websearch_to_tsquery('russian', $2) || 
				websearch_to_tsquery('english', $2) ||
				websearch_to_tsquery('simple', $2)
			) AS query
		) AS q
		WHERE u.search_vector @@ q.query
		ORDER BY rank DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, filter.Limit, filter.Query)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	var results []UserInfo
	var avatarURL sql.NullString

	for rows.Next() {
		var user UserInfo
		var rank float64

		if err := rows.Scan(&user.UserID, &user.Username, &avatarURL, &user.ReputationScore, &rank); err != nil {
			s.logger.WithError(err).Warn("scan user row")
			continue
		}

		if avatarURL.Valid {
			user.AvatarURL = avatarURL.String
		}
		results = append(results, user)
	}

	return results, rows.Err()
}
