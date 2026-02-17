package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Storage struct {
	db    *sqlx.DB
	close func() error
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
		close: func() error {
			return fmt.Errorf("close: %v", db.Close())
		},
	}
}

func (s *Storage) CreateUser(_ context.Context, filter SignUpFilter) (UserInfo, error) {
	query := `INSERT INTO member (username, email, password_hash)
            VALUES ($1, $2, $3) RETURNING id, role;`

	row := s.db.QueryRow(query, filter.Username, filter.Email, filter.HashedPassword)

	userInfo := UserInfo{
		Username: filter.Username,
		Email:    filter.Email,
	}

	err := row.Scan(&userInfo.UserID, &userInfo.Role)
	if err != nil {
		// Проверяем, была ли ошибка по уникальности.
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return UserInfo{}, fmt.Errorf("email already exists: %w", ErrUserAlreadyExists)
			}
		}
		return UserInfo{}, fmt.Errorf("failed to insert new user: %w", err)
	}

	return userInfo, nil
}
