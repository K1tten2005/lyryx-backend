package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserDoesntExist   = errors.New("user does not exist")
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
	query := `INSERT INTO users (username, email, password_hash)
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

func (s *Storage) GetHashedPasswordByEmail(_ context.Context, email string) (string, error) {
	query := `SELECT password_hash FROM users WHERE email = $1;`

	row := s.db.QueryRow(query, email)
	var hashedPassword string

	err := row.Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("failed to find user: %w", ErrUserDoesntExist)
		}
		return "", fmt.Errorf("failed to get hashed password^ %v", err)
	}

	return hashedPassword, nil
}

func (s *Storage) GetUserInfoByEmail(_ context.Context, email string) (UserInfo, error) {
	query := `SELECT
				id,
				username,
				reputation_score,
				role
			FROM users
			WHERE email = $1;
			`
	userInfo := UserInfo{
		Email: email,
	}

	row := s.db.QueryRow(query, email)

	err := row.Scan(&userInfo.UserID, &userInfo.Username, &userInfo.ReputationScore, &userInfo.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, fmt.Errorf("failed to find user: %w", ErrUserDoesntExist)
		}
		return UserInfo{}, fmt.Errorf("failed to get user info: %v", err)
	}

	return userInfo, nil
}

func (s *Storage) SetNewRefreshToken(_ context.Context, filter SetNewRefreshTokenFilter) error {
	query := `UPDATE users
				SET refresh_token = $1
				WHERE email = $2;`

	_, err := s.db.Exec(query, filter.RefreshToken, filter.Email)
	if err != nil {
		return fmt.Errorf("failed to set new refresh token: %v", err)
	}
	return nil
}

func (s *Storage) GetUserByRefreshToken(_ context.Context, refreshToken string) (UserInfo, error) {
	query := `SELECT
				id,
				email,
				username,
				role,
				reputation_score
			FROM users
          	WHERE refresh_token = $1;`

	row := s.db.QueryRow(query, refreshToken)

	var userInfo UserInfo

	err := row.Scan(&userInfo.UserID, &userInfo.Email, &userInfo.Username, &userInfo.Role, &userInfo.ReputationScore)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to get user: %v", err)
	}

	return userInfo, nil
}
