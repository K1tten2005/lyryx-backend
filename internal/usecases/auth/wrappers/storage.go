package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/auth/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/storage"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserDoesntExist   = errors.New("user does not exist")
)

type storage interface {
	CreateUser(_ context.Context, filter storageDto.SignUpFilter) (storageDto.UserInfo, error)
	GetHashedPasswordByEmail(_ context.Context, email string) (string, error)
	GetUserInfoByEmail(_ context.Context, email string) (storageDto.UserInfo, error)
	SetNewRefreshToken(_ context.Context, filter storageDto.SetNewRefreshTokenFilter) error
	GetUserByRefreshToken(_ context.Context, refrteshToken string) (storageDto.UserInfo, error)
}

type Storage struct {
	storage storage
}

func NewStorage(storage storage) *Storage {
	return &Storage{storage: storage}
}

func (s *Storage) CreateUser(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error) {
	filter := storageDto.SignUpFilter{
		Username:       opts.Username,
		Email:          opts.Email,
		HashedPassword: opts.Password,
	}

	userInfo, err := s.storage.CreateUser(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrUserAlreadyExists) {
			return dto.UserInfo{}, fmt.Errorf("create user: %w", ErrUserAlreadyExists)
		}
		return dto.UserInfo{}, fmt.Errorf("create user: %v", err)
	}

	return dto.UserInfo{
		UserID:          userInfo.UserID,
		Email:           userInfo.Email,
		Username:        userInfo.Username,
		ReputationScore: userInfo.ReputationScore,
		Role:            userInfo.Role,
	}, nil
}

func (s *Storage) GetHashedPasswordByEmail(ctx context.Context, email string) (string, error) {
	hashedPassword, err := s.storage.GetHashedPasswordByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storageDto.ErrUserDoesntExist) {
			return "", fmt.Errorf("get hashed password by email: %w", ErrUserDoesntExist)
		}
		return "", fmt.Errorf("get hashed password by email: %v", err)
	}
	return hashedPassword, nil
}

func (s *Storage) GetUserInfoByEmail(ctx context.Context, email string) (dto.UserInfo, error) {
	userInfo, err := s.storage.GetUserInfoByEmail(ctx, email)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("get user info by email: %v", err)
	}
	return dto.UserInfo{
		UserID:          userInfo.UserID,
		Email:           userInfo.Email,
		Username:        userInfo.Username,
		ReputationScore: userInfo.ReputationScore,
		Role:            userInfo.Role,
	}, nil
}

func (s *Storage) SetNewRefreshToken(ctx context.Context, opts dto.SetNewRefreshTokenOpts) error {
	filter := storageDto.SetNewRefreshTokenFilter{
		Email:        opts.Email,
		RefreshToken: opts.RefreshToken,
	}
	err := s.storage.SetNewRefreshToken(ctx, filter)
	if err != nil {
		return fmt.Errorf("set new refresh token: %v", err)
	}
	return nil
}

func (s *Storage) SignOut(ctx context.Context, opts dto.SignOutOpts) error {
	filter := storageDto.SetNewRefreshTokenFilter{
		Email:        opts.Email,
		RefreshToken: "",
	}

	err := s.storage.SetNewRefreshToken(ctx, filter)
	if err != nil {
		return fmt.Errorf("sign out: %v", err)
	}

	return nil
}

func (s *Storage) GetUserByRefreshToken(ctx context.Context, refreshToken string) (dto.UserInfo, error) {
	userInfo, err := s.storage.GetUserByRefreshToken(ctx, refreshToken)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("get user by refresh token: %v", err)
	}

	userInfoUC := dto.UserInfo{
		UserID:          userInfo.UserID,
		Email:           userInfo.Email,
		Username:        userInfo.Username,
		ReputationScore: userInfo.ReputationScore,
		Role:            userInfo.Role,
	}
	return userInfoUC, nil
}
