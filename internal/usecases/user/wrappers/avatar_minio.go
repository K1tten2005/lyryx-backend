package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/user/storage"
)

var (
	ErrInvalidAvatarType = errors.New("avatar must be a valid png/jpeg image")
	ErrAvatarTooLarge    = errors.New("avatar file is too large (max 5MB)")
)

type avatarStorage interface {
	UploadAvatar(ctx context.Context, filter storageDto.UploadAvatarFilter) (string, error)
}

type AvatarGetter struct {
	avatarStorage avatarStorage
}

func NewAvatarGetter(avatarStorage avatarStorage) *AvatarGetter {
	return &AvatarGetter{
		avatarStorage: avatarStorage,
	}
}

func (a *AvatarGetter) UploadAvatar(ctx context.Context, opts dto.UploadAvatarOpts) (string, error) {
	filter := storageDto.UploadAvatarFilter{
		UserID:     opts.UserID,
		AvatarFile: opts.AvatarFile,
	}
	
	avatarUrl, err := a.avatarStorage.UploadAvatar(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrInvalidAvatarType) {
			return "", fmt.Errorf("upload avatar: %w", ErrInvalidAvatarType)
		}
		if errors.Is(err, storageDto.ErrAvatarTooLarge) {
			return "", fmt.Errorf("upload avatar: %w", ErrAvatarTooLarge)
		}
		return "", fmt.Errorf("upload avatar: %v", err)
	}
	return avatarUrl, nil
}
