package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/artist/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/artist/storage"
)

var (
	ErrArtistNotFound = errors.New("artist not found")
	ErrNameTaken      = errors.New("artist name already exists")
)

type storage interface {
	GetArtistByID(_ context.Context, artistID int) (storageDto.Artist, error)
	CreateArtist(_ context.Context, filter storageDto.CreateArtistFilter) (storageDto.Artist, error)
	UpdateArtistInfo(_ context.Context, filter storageDto.UpdateArtistInfoFilter) (storageDto.Artist, error)
	UpdateArtistAvatar(_ context.Context, filter storageDto.UpdateArtistAvatarFilter) error
}

type Storage struct {
	storage storage
}

func NewStorage(storage storage) *Storage {
	return &Storage{
		storage: storage,
	}
}

func (s *Storage) GetArtistByID(ctx context.Context, artistID int) (dto.Artist, error) {
	artist, err := s.storage.GetArtistByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, storageDto.ErrArtistNotFound) {
			return dto.Artist{}, ErrArtistNotFound
		}
		return dto.Artist{}, fmt.Errorf("get artist by id: %v", err)
	}

	return dto.Artist{
		ArtistID:  artist.ArtistID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}, nil
}

func (s *Storage) PostArtist(ctx context.Context, opts dto.PostArtistOpts) (dto.Artist, error) {
	filter := storageDto.CreateArtistFilter{
		Name: opts.Name,
		Bio:  opts.Bio,
	}
	artist, err := s.storage.CreateArtist(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrNameTaken) {
			return dto.Artist{}, fmt.Errorf("post artist: %w", ErrNameTaken)
		}
		return dto.Artist{}, fmt.Errorf("get artist by id: %v", err)
	}
	artistUC := dto.Artist{
		ArtistID:  artist.ArtistID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}

	return artistUC, nil
}

func (s *Storage) PatchUpdateArtist(ctx context.Context, opts dto.PatchUpdateArtistOpts) (dto.Artist, error) {
	filter := storageDto.UpdateArtistInfoFilter{
		ArtistID: opts.ArtistID,
		Name:     opts.Name,
		Bio:      opts.Bio,
	}

	artist, err := s.storage.UpdateArtistInfo(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrArtistNotFound) {
			return dto.Artist{}, fmt.Errorf("patch update artist: %w", ErrArtistNotFound)
		}
		if errors.Is(err, storageDto.ErrNameTaken) {
			return dto.Artist{}, fmt.Errorf("patch update artist: %w", ErrNameTaken)
		}
		return dto.Artist{}, fmt.Errorf("patch update artist: %v", err)
	}

	return dto.Artist{
		ArtistID:  artist.ArtistID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}, nil
}

func (s *Storage) PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error {
	filter := storageDto.UpdateArtistAvatarFilter{
		ArtistID:  opts.ArtistID,
		AvatarURL: opts.AvatarURL,
	}
	err := s.storage.UpdateArtistAvatar(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrArtistNotFound) {
			return fmt.Errorf("patch update avatar: %w", ErrArtistNotFound)
		}
		return fmt.Errorf("patch update avatar: %v", err)
	}

	return nil
}
