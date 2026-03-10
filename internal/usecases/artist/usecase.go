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
	ErrNameTaken      = errors.New("artist name already exists")
	ErrInvalidAvatarType = errors.New("avatar must be a valid png/jpeg image")
	ErrAvatarTooLarge    = errors.New("avatar file is too large (max 5MB)")
)

type storage interface {
	GetArtistByID(ctx context.Context, artistID int) (dto.Artist, error)
	PostArtist(ctx context.Context, opts dto.PostArtistOpts) (dto.Artist, error)
}

type Usecase struct {
	storage storage

	logger *logrus.Logger
}

func NewUsecase(
	storage storage,

	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage: storage,
		logger:  logger,
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
