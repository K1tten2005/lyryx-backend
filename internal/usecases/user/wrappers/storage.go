package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/user/storage"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type storage interface {
	GetUserByID(_ context.Context, userID int) (storageDto.User, error)
}

type Storage struct {
	storage storage
}

func NewStorage(storage storage) *Storage {
	return &Storage{storage: storage}
}

func (s *Storage) GetUserByID(ctx context.Context, userID int) (dto.User, error) {
	storageProfile, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, storageDto.ErrUserNotFound) {
			return dto.User{}, ErrUserNotFound
		}
		return dto.User{}, fmt.Errorf("get user by id: %v", err)
	}

	return dto.User{
		UserID:          storageProfile.UserID,
		Email:           storageProfile.Email,
		Username:        storageProfile.Username,
		ReputationScore: storageProfile.ReputationScore,
		Role:            storageProfile.Role,
	}, nil
}
