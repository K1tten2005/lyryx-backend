package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/artist/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/artist/storage"
)

var (
	ErrInvalidAvatarType = errors.New("avatar must be a valid png/jpeg image")
	ErrAvatarTooLarge    = errors.New("avatar file is too large (max 5MB)")
)

type artistAvatarStorage interface {
	UploadAvatar(ctx context.Context, filter storageDto.UploadAvatarFilter) (string, error)
}

type ArtistAvatarStorage struct {
	artistAvatarStorage artistAvatarStorage
}

func NewArtistAvatarStorage(artistAvatarStorage artistAvatarStorage) *ArtistAvatarStorage {
	return &ArtistAvatarStorage{
		artistAvatarStorage: artistAvatarStorage,
	}
}

func (ua *ArtistAvatarStorage) UploadAvatar(ctx context.Context, opts dto.UploadAvatarOpts) (string, error) {
	filter := storageDto.UploadAvatarFilter{
		ArtistID:   opts.ArtistID,
		AvatarFile: opts.AvatarFile,
	}

	avatarUrl, err := ua.artistAvatarStorage.UploadAvatar(ctx, filter)
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
