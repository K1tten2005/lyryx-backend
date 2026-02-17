package wrappers

import (
	"context"
	"errors"
	"fmt"
	"lyryx-backend/internal/usecases/auth/dto"
	storageDto "lyryx-backend/internal/usecases/auth/storage"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type storage interface {
	CreateUser(_ context.Context, filter storageDto.SignUpFilter) (storageDto.UserInfo, error)
}

type Storage struct {
	storage storage
}

func NewStorage(storage storage) *Storage {
	return &Storage{storage: storage}
}

func (s *Storage) PostSignUp(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error) {
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
		UserID:   userInfo.UserID,
		Email:    userInfo.Email,
		Username: userInfo.Username,
		Role:     userInfo.Role,
	}, nil
}
