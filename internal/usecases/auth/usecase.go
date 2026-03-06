package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/auth/dto"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/auth/wrappers"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserDoesntExist   = errors.New("user does not exist")
	ErrInvalidPassword   = errors.New("invalid password")
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
	CreateUser(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error)
	GetHashedPasswordByEmail(ctx context.Context, email string) (string, error)
	GetUserInfoByEmail(ctx context.Context, email string) (dto.UserInfo, error)
	SetNewRefreshToken(ctx context.Context, opts dto.SetNewRefreshTokenOpts) error
	SignOut(ctx context.Context, opts dto.SignOutOpts) error
	GetUserByRefreshToken(ctx context.Context, refreshToken string) (dto.UserInfo, error)
}

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
	userInfo, err := u.storage.CreateUser(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserAlreadyExists) {
			return dto.UserInfo{}, fmt.Errorf("post sign up: %w", ErrUserAlreadyExists)
		}
		return dto.UserInfo{}, fmt.Errorf("post sign up: %v", err)
	}

	return userInfo, nil
}

func (u *Usecase) PostSignIn(ctx context.Context, opts dto.SignInOpts) (dto.UserInfo, error) {
	// 1. Идем в бд за хешированным паролем.
	hashedPasswordFromDB, err := u.storage.GetHashedPasswordByEmail(ctx, opts.Email)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserDoesntExist) {
			return dto.UserInfo{}, fmt.Errorf("get user by email: %w", ErrUserDoesntExist)
		}
		return dto.UserInfo{}, fmt.Errorf("user not found or database error: %v", err)
	}

	// 2. Сравниваем хеши.
	err = bcrypt.CompareHashAndPassword([]byte(hashedPasswordFromDB), []byte(opts.Password))
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("compare password: %w", ErrInvalidPassword)
	}

	// 3. Берем информацию пользователя.
	userInfo, err := u.storage.GetUserInfoByEmail(ctx, opts.Email)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("post sign in: %v", err)
	}

	return userInfo, nil
}

func (u *Usecase) SetNewRefreshToken(ctx context.Context, opts dto.SetNewRefreshTokenOpts) error {
	err := u.storage.SetNewRefreshToken(ctx, opts)
	if err != nil {
		return fmt.Errorf("set new refresh token: %v", err)
	}
	return nil
}

func (u *Usecase) PostSignOut(ctx context.Context, opts dto.SignOutOpts) error {
	err := u.storage.SignOut(ctx, opts)
	if err != nil {
		return fmt.Errorf("post sign out: %v", err)
	}

	return nil
}

func (u *Usecase) GetUserByRefreshToken(ctx context.Context, refreshToken string) (dto.UserInfo, error) {
	userInfo, err := u.storage.GetUserByRefreshToken(ctx, refreshToken)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("get user by refresh token: %v", err)
	}
	return userInfo, nil
}
