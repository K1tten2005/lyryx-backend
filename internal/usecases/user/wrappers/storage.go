package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/user/storage"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameTaken      = errors.New("username already exists")
)

type storage interface {
	GetUserByID(_ context.Context, userID int) (storageDto.User, error)
	PatchUpdateUser(_ context.Context, opts dto.PatchUpdateUserOpts) error
	PatchUpdateAvatar(_ context.Context, opts dto.PatchUpdateAvatarOpts) error
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
		Bio:             storageProfile.Bio,
		AvatarURL:       storageProfile.AvatarURL,
		ReputationScore: storageProfile.ReputationScore,
		Role:            storageProfile.Role,
	}, nil
}

func (s *Storage) PatchUpdateUser(ctx context.Context, opts dto.PatchUpdateUserOpts) error {
	err := s.storage.PatchUpdateUser(ctx, opts)
	if err != nil {
		if errors.Is(err, storageDto.ErrUserNotFound) {
			return ErrUserNotFound
		}
		if errors.Is(err, storageDto.ErrEmailAlreadyExists) {
			return ErrEmailAlreadyExists
		}
		if errors.Is(err, storageDto.ErrUsernameTaken) {
			return ErrUsernameTaken
		}
		return fmt.Errorf("patch update user: %v", err)
	}

	return nil
}

func (s *Storage) PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error {
	err := s.storage.PatchUpdateAvatar(ctx, opts)
	if err != nil {
		if errors.Is(err, storageDto.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("patch update avatar: %v", err)
	}

	return nil
}
