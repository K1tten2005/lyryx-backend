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
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameTaken      = errors.New("username already exists")
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

func (s *Storage) GetUserByID(_ context.Context, userID int) (User, error) {
	query := `
        SELECT
			id,
			username,
			email,
			bio,
			avatar_url,
			reputation_score,
			role
        FROM users
        WHERE id = $1
    `

	row := s.db.QueryRow(query, userID)

	var user User
	var bio sql.NullString
	var avatarURL sql.NullString
	if err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&bio,
		&avatarURL,
		&user.ReputationScore,
		&user.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("query: %v", err)
	}

	if bio.Valid {
		user.Bio = bio.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return user, nil
}

func (s *Storage) UpdateUserInfo(_ context.Context, filter UpdateUserInfoFilter) error {
	query := `
		UPDATE users
		SET
			email = COALESCE($2, email),
			username = COALESCE($3, username),
			bio = COALESCE($4, bio),
			password_hash = COALESCE($5, password_hash)
		WHERE id = $1
	`

	res, err := s.db.Exec(query, filter.UserID, filter.Email, filter.Username, filter.Bio, filter.Password)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return ErrEmailAlreadyExists
		}
		if isUniqueViolation(err, "users_username_key") {
			return ErrUsernameTaken
		}
		return fmt.Errorf("exec patch update user: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Storage) UpdateUserAvatar(_ context.Context, filter UpdateUserAvatarFilter) error {
	query := `
		UPDATE users
		SET avatar_url = $2
		WHERE id = $1
	`

	res, err := s.db.Exec(query, filter.UserID, filter.AvatarURL)
	if err != nil {
		return fmt.Errorf("exec patch update avatar: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
