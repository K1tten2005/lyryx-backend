package song

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/song/dto"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/song/wrappers"
	"github.com/sirupsen/logrus"
)

var (
	ErrSongNotFound     = errors.New("song not found")
	ErrInvalidCoverType = errors.New("cover must be a valid png/jpeg image")
	ErrCoverTooLarge    = errors.New("cover file is too large (max 5MB)")
)

type storage interface {
	GetSongByID(ctx context.Context, artistID int) (dto.SongInfo, error)
	PostSong(ctx context.Context, opts dto.PostSongOpts) (dto.SongInfo, error)
	PatchUpdateSong(ctx context.Context, opts dto.PatchUpdateSongOpts) (dto.SongInfo, error)
	PatchUpdateCover(ctx context.Context, opts dto.PatchUpdateCoverOpts) error
}

type songCoverUploader interface {
	UploadCover(ctx context.Context, opts dto.UploadCoverOpts) (string, error)
}

type Usecase struct {
	storage           storage
	songCoverUploader songCoverUploader

	logger *logrus.Logger
}

func NewUsecase(
	storage storage,
	songCoverUploader songCoverUploader,

	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage:           storage,
		songCoverUploader: songCoverUploader,
		logger:            logger,
	}
}

func (u *Usecase) GetSongByID(ctx context.Context, songID int) (dto.SongInfo, error) {
	song, err := u.storage.GetSongByID(ctx, songID)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return dto.SongInfo{}, ErrSongNotFound
		}
		return dto.SongInfo{}, fmt.Errorf("get song by id: %v", err)
	}

	return song, nil
}

func (u *Usecase) PostSong(ctx context.Context, opts dto.PostSongOpts) (dto.SongInfo, error) {
	song, err := u.storage.PostSong(ctx, opts)
	if err != nil {
		return dto.SongInfo{}, fmt.Errorf("post song: %v", err)
	}

	return song, nil
}

func (u *Usecase) PatchUpdateSong(ctx context.Context, opts dto.PatchUpdateSongOpts) (dto.SongInfo, error) {
	song, err := u.storage.PatchUpdateSong(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return dto.SongInfo{}, fmt.Errorf("patch update song: %w", ErrSongNotFound)
		}
		return dto.SongInfo{}, fmt.Errorf("patch update song: %v", err)
	}

	return song, nil
}

func (u *Usecase) PatchUpdateCover(ctx context.Context, opts dto.UploadCoverOpts) (string, error) {
	// 1. Проверяем, что пользователь существует.
	_, err := u.storage.GetSongByID(ctx, opts.SongID)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return "", fmt.Errorf("patch update cover: %w", ErrSongNotFound)
		}
		return "", fmt.Errorf("patch update song: %v", err)
	}

	// 2. Загружаем аватар в minio.
	coverUrl, err := u.songCoverUploader.UploadCover(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrInvalidCoverType) {
			return "", fmt.Errorf("patch update cover: %v", ErrInvalidCoverType)
		}
		if errors.Is(err, wrappers.ErrCoverTooLarge) {
			return "", fmt.Errorf("patch update cover: %v", ErrCoverTooLarge)
		}
		return "", fmt.Errorf("patch update cover: %v", err)
	}

	// 2. Обновляем ссылку на аватар в бд.
	err = u.storage.PatchUpdateCover(ctx, dto.PatchUpdateCoverOpts{
		SongID:    opts.SongID,
		CoverURL:  coverUrl,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return "", ErrSongNotFound
		}
		return "", fmt.Errorf("patch update cover: %v", err)
	}

	return coverUrl, nil
}
