package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/song/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/song/storage"
)

var (
	ErrInvalidCoverType = errors.New("cover must be a valid png/jpeg image")
	ErrCoverTooLarge    = errors.New("cover file is too large (max 5MB)")
)

type songCoverStorage interface {
	UploadCover(ctx context.Context, filter storageDto.UploadCoverFilter) (string, error)
}

type SongCoverStorage struct {
	songCoverStorage songCoverStorage
}

func NewSongCoverStorage(songCoverStorage songCoverStorage) *SongCoverStorage {
	return &SongCoverStorage{
		songCoverStorage: songCoverStorage,
	}
}

func (sc *SongCoverStorage) UploadCover(ctx context.Context, opts dto.UploadCoverOpts) (string, error) {
	filter := storageDto.UploadCoverFilter{
		SongID:    opts.SongID,
		CoverFile: opts.CoverFile,
	}

	avatarUrl, err := sc.songCoverStorage.UploadCover(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrInvalidCoverType) {
			return "", fmt.Errorf("upload cover: %w", ErrInvalidCoverType)
		}
		if errors.Is(err, storageDto.ErrCoverTooLarge) {
			return "", fmt.Errorf("upload cover: %w", ErrCoverTooLarge)
		}
		return "", fmt.Errorf("upload cover: %v", err)
	}
	return avatarUrl, nil
}
