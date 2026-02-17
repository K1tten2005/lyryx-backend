package auth

import (
	"context"
	"errors"
	"fmt"
	"lyryx-backend/internal/usecases/auth/dto"
	"lyryx-backend/internal/usecases/auth/wrappers"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

const (
	hashCost = 12 // Стоимость хеширования (рекомендуется от 10 до 14).
)

// Выносим в интерфейс, чтобы можно было создать моки.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
}

type BcryptHasher struct{}

func (b *BcryptHasher) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

type storage interface {
PostSignUp(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error)}

type Usecase struct {
	storage storage
	hasher  PasswordHasher
}

func NewUsecase(storage storage, hasher PasswordHasher) *Usecase {
	return &Usecase{
		storage: storage,
		hasher:  hasher,
	}
}

func (u *Usecase) PostSignUp(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error) {
	// 1. Хешируем пароль.
	hashedPassword, err := u.hasher.HashPassword(opts.Password)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("hash password: %v", err)
	}

	// 2. Устанавливаем захешированный пароль.
	opts.Password = hashedPassword

	// 3. Сохраняем пользователя.
	userInfo, err := u.storage.PostSignUp(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserAlreadyExists) {
			return dto.UserInfo{}, fmt.Errorf("post sign up: %w", ErrUserAlreadyExists)
		}
		return dto.UserInfo{}, fmt.Errorf("post sign up: %v", err)
	}

	return userInfo, nil
}
