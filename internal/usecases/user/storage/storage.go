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
	ErrUserNotFound = errors.New("user not found")
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
			reputation_score,
			role
        FROM users
        WHERE id = $1
    `

	row := s.db.QueryRow(query, userID)

	var user User
	if err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.ReputationScore,
		&user.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("query: %v", err)
	}

	return user, nil
}
