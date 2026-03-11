package artist

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/artist/dto"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/artist/wrappers"
	"github.com/sirupsen/logrus"
)

var (
	ErrArtistNotFound    = errors.New("artist not found")
	ErrNameTaken         = errors.New("artist name already exists")
	ErrInvalidAvatarType = errors.New("avatar must be a valid png/jpeg image")
	ErrAvatarTooLarge    = errors.New("avatar file is too large (max 5MB)")
)

type storage interface {
	GetArtistByID(ctx context.Context, artistID int) (dto.Artist, error)
	PostArtist(ctx context.Context, opts dto.PostArtistOpts) (dto.Artist, error)
	PatchUpdateArtist(ctx context.Context, opts dto.PatchUpdateArtistOpts) (dto.Artist, error)
	PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error
}

type artistAvatarUploader interface {
	UploadAvatar(ctx context.Context, opts dto.UploadAvatarOpts) (string, error)
}

type Usecase struct {
	storage              storage
	artistAvatarUploader artistAvatarUploader

	logger *logrus.Logger
}

func NewUsecase(
	storage storage,
	artistAvatarUploader artistAvatarUploader,

	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage:              storage,
		artistAvatarUploader: artistAvatarUploader,
		logger:               logger,
	}
}

func (u *Usecase) GetArtistByID(ctx context.Context, artistID int) (dto.Artist, error) {
	user, err := u.storage.GetArtistByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, wrappers.ErrArtistNotFound) {
			return dto.Artist{}, ErrArtistNotFound
		}
		return dto.Artist{}, fmt.Errorf("get artist by id: %v", err)
	}

	return user, nil
}

func (u *Usecase) PostArtist(ctx context.Context, opts dto.PostArtistOpts) (dto.Artist, error) {
	artist, err := u.storage.PostArtist(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrNameTaken) {
			return dto.Artist{}, fmt.Errorf("post artist: %w", ErrNameTaken)
		}
		return dto.Artist{}, fmt.Errorf("post artist: %v", err)
	}

	return artist, nil
}

func (u *Usecase) PatchUpdateArtist(ctx context.Context, opts dto.PatchUpdateArtistOpts) (dto.Artist, error) {
	artist, err := u.storage.PatchUpdateArtist(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrNameTaken) {
			return dto.Artist{}, fmt.Errorf("patch update artist: %w", ErrNameTaken)
		}
		if errors.Is(err, wrappers.ErrArtistNotFound) {
			return dto.Artist{}, fmt.Errorf("patch update artist: %w", ErrArtistNotFound)
		}
		return dto.Artist{}, fmt.Errorf("patch update artist: %v", err)
	}

	return artist, nil
}

func (u *Usecase) PatchUpdateAvatar(ctx context.Context, opts dto.UploadAvatarOpts) (string, error) {
	// 1. Проверяем, что пользователь существует.
	_, err := u.storage.GetArtistByID(ctx, opts.ArtistID)
	if err != nil {
		if errors.Is(err, wrappers.ErrArtistNotFound) {
			return "", fmt.Errorf("patch update avatar: %w", ErrArtistNotFound)
		}
		return "", fmt.Errorf("patch update avatar: %v", err)
	}

	// 2. Загружаем аватар в minio.
	avatarUrl, err := u.artistAvatarUploader.UploadAvatar(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrInvalidAvatarType) {
			return "", fmt.Errorf("patch update avatar: %v", ErrInvalidAvatarType)
		}
		if errors.Is(err, wrappers.ErrAvatarTooLarge) {
			return "", fmt.Errorf("patch update avatar: %v", ErrAvatarTooLarge)
		}
		return "", fmt.Errorf("patch update avatar: %v", err)
	}

	// 2. Обновляем ссылку на аватар в бд.
	err = u.storage.PatchUpdateAvatar(ctx, dto.PatchUpdateAvatarOpts{
		ArtistID:  opts.ArtistID,
		AvatarURL: avatarUrl,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrArtistNotFound) {
			return "", ErrArtistNotFound
		}
		return "", fmt.Errorf("patch update avatar: %v", err)
	}

	return avatarUrl, nil
}
