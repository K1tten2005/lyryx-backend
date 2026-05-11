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
	ErrAnnotationsNotFound = errors.New("annotations not found")
	ErrSongNotFound        = errors.New("song not found")
	ErrAnnotationNotFound  = errors.New("annotation not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrAnnotationOverlap   = errors.New("annotation overlaps with an existing one")
)

type Storage struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewStorage(db *sqlx.DB, logger *logrus.Logger) *Storage {
	return &Storage{
		db:     db,
		logger: logger,
	}
}

// GetAnnotations получает список аннотаций для песни
func (s *Storage) GetAnnotations(ctx context.Context, filter GetAnnotationsFilter) ([]AnnotationInfo, error) {
	query := `
        SELECT
            ann.id, 
            ann.content, 
            ann.start_index, 
            ann.end_index, 
            ann.rating, 
            ann.created_at,
            ann.updated_at,
            u.id as user_id, 
            u.username, 
            u.avatar_url, 
            u.reputation_score,
            s.id as song_id,
            s.title,
            s.cover_url,
            art.id as artist_id,
            art.name as artist_name,
            v.vote_value as my_vote
        FROM annotation ann
        JOIN users u ON ann.author_id = u.id
        JOIN song s ON ann.song_id = s.id
        JOIN artist art ON s.artist_id = art.id
        LEFT JOIN annotation_vote v 
    		ON ann.id = v.annotation_id 
			AND v.user_id = $1
        WHERE ann.song_id = $2
        ORDER BY ann.start_index ASC, ann.created_at ASC
    `

	rows, err := s.db.QueryContext(ctx, query, filter.UserID, filter.SongID)
	if err != nil {
		return nil, fmt.Errorf("query annotations: %w", err)
	}
	defer rows.Close()

	var annotations []AnnotationInfo
	for rows.Next() {
		var a AnnotationInfo
		var myVote sql.NullInt64
		var avatar sql.NullString

		err := rows.Scan(
			&a.ID, &a.Content, &a.StartIndex, &a.EndIndex, &a.Rating, &a.CreatedAt, &a.UpdatedAt,
			&a.User.UserID, &a.User.Username, &avatar, &a.User.ReputationScore,
			&a.Song.ID, &a.Song.Title, &a.Song.CoverURL,
			&a.Song.Artist.ID, &a.Song.Artist.Name,
			&myVote,
		)
		if err != nil {
			return nil, fmt.Errorf("scan annotation: %w", err)
		}

		if avatar.Valid {
			a.User.AvatarURL = avatar.String
		}
		if myVote.Valid {
			val := int(myVote.Int64)
			a.MyVote = &val
		}

		annotations = append(annotations, a)
	}
	if len(annotations) == 0 {
		return nil, ErrAnnotationsNotFound
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return annotations, nil
}

// GetAnnotationByID получает одну аннотацию
func (s *Storage) GetAnnotationByID(ctx context.Context, filter GetAnnotationByIDFilter) (AnnotationInfo, error) {
	query := `
        SELECT
            ann.id, 
            ann.content, 
            ann.start_index, 
            ann.end_index, 
            ann.rating, 
            ann.created_at,
            ann.updated_at,
            u.id as user_id, 
            u.username, 
            u.avatar_url, 
            u.reputation_score,
            s.id as song_id,
            s.title,
            s.cover_url,
            art.id as artist_id,
            art.name as artist_name,
            v.vote_value as my_vote
        FROM annotation ann
        JOIN users u ON ann.author_id = u.id
        JOIN song s ON ann.song_id = s.id
        JOIN artist art ON s.artist_id = art.id
        LEFT JOIN annotation_vote v 
    		ON ann.id = v.annotation_id 
   			AND v.user_id = $1
        WHERE ann.id = $2
    `

	var a AnnotationInfo
	var myVote sql.NullInt64
	var avatar sql.NullString

	err := s.db.QueryRowContext(ctx, query, filter.UserID, filter.AnnotationID).Scan(
		&a.ID, &a.Content, &a.StartIndex, &a.EndIndex, &a.Rating, &a.CreatedAt, &a.UpdatedAt,
		&a.User.UserID, &a.User.Username, &avatar, &a.User.ReputationScore,
		&a.Song.ID, &a.Song.Title, &a.Song.CoverURL,
		&a.Song.Artist.ID, &a.Song.Artist.Name,
		&myVote,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AnnotationInfo{}, ErrAnnotationNotFound
		}
		return AnnotationInfo{}, fmt.Errorf("get annotation by id: %w", err)
	}

	if avatar.Valid {
		a.User.AvatarURL = avatar.String
	}
	if myVote.Valid {
		val := int(myVote.Int64)
		a.MyVote = &val
	}

	return a, nil
}

// CreateAnnotation создает новую аннотацию
func (s *Storage) CreateAnnotation(ctx context.Context, filter CreateAnnotationFilter) (AnnotationInfo, error) {
	// 1. ПРОВЕРКА НА ПЕРЕСЕЧЕНИЕ
	// Ищем аннотацию для этой же песни, где (существующий_start < новый_end) И (существующий_end > новый_start)
	overlapQuery := `
        SELECT EXISTS (
            SELECT 1 FROM annotation 
            WHERE song_id = $1 
              AND start_index < $2 
              AND end_index > $3
        )
    `
	var hasOverlap bool
	err := s.db.GetContext(ctx, &hasOverlap, overlapQuery, filter.SongID, filter.EndIndex, filter.StartIndex)
	if err != nil {
		return AnnotationInfo{}, fmt.Errorf("check overlap: %w", err)
	}

	if hasOverlap {
		return AnnotationInfo{}, ErrAnnotationOverlap
	}

	query := `
        INSERT INTO annotation (author_id, song_id, content, start_index, end_index, rating, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, 0, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `

	var newAnn AnnotationInfo
	// Заполняем базовые данные, которые у нас есть
	newAnn.User.UserID = filter.AuthorID
	newAnn.Song.ID = filter.SongID

	err = s.db.QueryRowContext(ctx, query, filter.AuthorID, filter.SongID, filter.Content, filter.StartIndex, filter.EndIndex).Scan(
		&newAnn.ID,
		&newAnn.CreatedAt,
		&newAnn.UpdatedAt,
	)

	if err != nil {
		return AnnotationInfo{}, fmt.Errorf("create annotation: %w", err)
	}

	return s.GetAnnotationByID(ctx, GetAnnotationByIDFilter{
		AnnotationID: newAnn.ID,
		UserID:       &filter.AuthorID,
	})
}

// UpdateAnnotation обновляет контент аннотации
func (s *Storage) UpdateAnnotation(ctx context.Context, filter UpdateAnnotationFilter) (AnnotationInfo, error) {
	query := `
        UPDATE annotation 
        SET content = $1, updated_at = NOW()
        WHERE id = $2
    `

	row := s.db.QueryRowContext(ctx, query, filter.Content, filter.AnnotationID)
	if row.Err() != nil {
		return AnnotationInfo{}, fmt.Errorf("update annotation: %w", row.Err())
	}

	return s.GetAnnotationByID(ctx, GetAnnotationByIDFilter{
		AnnotationID: filter.AnnotationID,
		UserID:       &filter.UserID,
	})
}

// DeleteAnnotation удаляет аннотацию
func (s *Storage) DeleteAnnotation(ctx context.Context, filter DeleteAnnotationFilter) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM annotation WHERE id = $1", filter.AnnotationID)
	if err != nil {
		return fmt.Errorf("delete annotation: %w", err)
	}
	return nil
}

// VoteAnnotation ставит голос
func (s *Storage) VoteAnnotation(ctx context.Context, filter VoteAnnotationFilter) (int, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Проверяем текущий голос
	var oldVoteValue sql.NullInt64
	err = tx.QueryRowContext(ctx, "SELECT vote_value FROM annotation_vote WHERE user_id = $1 AND annotation_id = $2", filter.UserID, filter.AnnotationID).Scan(&oldVoteValue)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("check existing vote: %w", err)
	}

	oldVal := 0
	if oldVoteValue.Valid {
		oldVal = int(oldVoteValue.Int64)
	}

	// Если голос не изменился, просто возвращаем текущий рейтинг
	if oldVal == filter.Value {
		var currentRating int
		err = tx.GetContext(ctx, &currentRating, "SELECT rating FROM annotation WHERE id = $1", filter.AnnotationID)
		if err != nil {
			return 0, err
		}
		tx.Commit()
		return currentRating, nil
	}

	// 2. Upsert голоса
	if oldVoteValue.Valid {
		_, err = tx.ExecContext(ctx, "UPDATE annotation_vote SET vote_value = $1, created_at = NOW() WHERE user_id = $2 AND annotation_id = $3", filter.Value, filter.UserID, filter.AnnotationID)
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO annotation_vote (user_id, annotation_id, vote_value, created_at) VALUES ($1, $2, $3, NOW())", filter.UserID, filter.AnnotationID, filter.Value)
	}

	if err != nil {
		return 0, fmt.Errorf("upsert vote_value: %w", err)
	}

	// 3. Пересчет рейтинга
	var newRating int
	err = tx.GetContext(ctx, &newRating, "SELECT COALESCE(SUM(vote_value), 0) FROM annotation_vote WHERE annotation_id = $1", filter.AnnotationID)
	if err != nil {
		return 0, fmt.Errorf("calc new rating: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE annotation SET rating = $1 WHERE id = $2", newRating, filter.AnnotationID)
	if err != nil {
		return 0, fmt.Errorf("update annotation rating: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return newRating, nil
}

// DeleteVote удаляет голос
func (s *Storage) DeleteVote(ctx context.Context, filter RemoveVoteFilter) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM annotation_vote WHERE user_id = $1 AND annotation_id = $2", filter.UserID, filter.AnnotationID)
	if err != nil {
		return err
	}

	var newRating int
	err = tx.GetContext(ctx, &newRating, "SELECT COALESCE(SUM(vote_value), 0) FROM annotation_vote WHERE annotation_id = $1", filter.AnnotationID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE annotation SET rating = $1 WHERE id = $2", newRating, filter.AnnotationID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// GetUserAnnotations получает аннотации пользователя
func (s *Storage) GetUserAnnotations(ctx context.Context, filter GetUserAnnotationsFilter) ([]AnnotationInfo, int, error) {
	query := `
        SELECT
            ann.id, 
            ann.content, 
            ann.start_index, 
            ann.end_index, 
            ann.rating, 
            ann.created_at,
            ann.updated_at,
            u.id as user_id, 
            u.username, 
            u.avatar_url, 
            u.reputation_score,
            s.id as song_id,
            s.title,
            s.cover_url,
			s.lyrics,
            art.id as artist_id,
            art.name as artist_name,
            v.vote_value as my_vote
        FROM annotation ann
        JOIN users u ON ann.author_id = u.id
        JOIN song s ON ann.song_id = s.id
        JOIN artist art ON s.artist_id = art.id
        LEFT JOIN annotation_vote v 
    		ON ann.id = v.annotation_id 
   			AND v.user_id = $1
        WHERE ann.author_id = $2
        ORDER BY ann.created_at DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := s.db.QueryContext(ctx, query, filter.CurrentUserID, filter.UserID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var annotations []AnnotationInfo
	for rows.Next() {
		var (
			a      AnnotationInfo
			myVote sql.NullInt64
			avatar sql.NullString
			lyrics string
		)

		err := rows.Scan(
			&a.ID, &a.Content, &a.StartIndex, &a.EndIndex, &a.Rating, &a.CreatedAt, &a.UpdatedAt,
			&a.User.UserID, &a.User.Username, &avatar, &a.User.ReputationScore,
			&a.Song.ID, &a.Song.Title, &a.Song.CoverURL, &lyrics,
			&a.Song.Artist.ID, &a.Song.Artist.Name, &myVote,
		)
		if err != nil {
			return nil, 0, err
		}

		runes := []rune(lyrics)
		a.Snippet = string(runes[a.StartIndex:a.EndIndex])

		if avatar.Valid {
			a.User.AvatarURL = avatar.String
		}
		if myVote.Valid {
			val := int(myVote.Int64)
			a.MyVote = &val
		}
		annotations = append(annotations, a)
	}

	var total int
	err = s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM annotation WHERE author_id = $1", filter.UserID)
	if err != nil {
		return nil, 0, err
	}

	return annotations, total, nil
}
