package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/song/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/song/storage"
)

var (
	ErrSongNotFound = errors.New("artist not found")
)

type storage interface {
	GetSongByID(_ context.Context, songID int) (storageDto.SongInfo, error)
	CreateSong(_ context.Context, filter storageDto.CreateSongFilter) (storageDto.SongInfo, error)
	UpdateSongInfo(_ context.Context, filter storageDto.UpdateSongInfoFilter) (storageDto.SongInfo, error)
	UpdateSongCover(_ context.Context, filter storageDto.UpdateSongCoverFilter) error
}

type Storage struct {
	storage storage
}

func NewStorage(storage storage) *Storage {
	return &Storage{
		storage: storage,
	}
}

func (s *Storage) GetSongByID(ctx context.Context, songID int) (dto.SongInfo, error) {
	song, err := s.storage.GetSongByID(ctx, songID)
	if err != nil {
		if errors.Is(err, storageDto.ErrSongNotFound) {
			return dto.SongInfo{}, ErrSongNotFound
		}
		return dto.SongInfo{}, fmt.Errorf("get song by id: %v", err)
	}

	return dto.SongInfo{
		SongID:      song.SongID,
		Title:       song.Title,
		Lyrics:      song.Lyrics,
		CoverURL:    song.CoverURL,
		ReleaseDate: song.ReleaseDate,
		Views:       song.Views,
		Artist: dto.Artist{
			ArtistID:  song.Artist.ArtistID,
			Name:      song.Artist.Name,
			Bio:       song.Artist.Bio,
			AvatarURL: song.Artist.AvatarURL,
		},
	}, nil
}

func (s *Storage) PostSong(ctx context.Context, opts dto.PostSongOpts) (dto.SongInfo, error) {
	filter := storageDto.CreateSongFilter{
		Title:       opts.Title,
		Lyrics:      opts.Lyrics,
		CoverURL:    opts.CoverURL,
		ReleaseDate: opts.ReleaseDate,
		ArtistID:    opts.ArtistID,
	}
	song, err := s.storage.CreateSong(ctx, filter)
	if err != nil {
		return dto.SongInfo{}, fmt.Errorf("get song by id: %v", err)
	}
	songUC := dto.SongInfo{
		SongID:      song.SongID,
		Title:       song.Title,
		Lyrics:      song.Lyrics,
		CoverURL:    song.CoverURL,
		ReleaseDate: song.ReleaseDate,
		Views:       song.Views,
		Artist: dto.Artist{
			ArtistID:  song.Artist.ArtistID,
			Name:      song.Artist.Name,
			Bio:       song.Artist.Bio,
			AvatarURL: song.Artist.AvatarURL,
		},
	}

	return songUC, nil
}

func (s *Storage) PatchUpdateSong(ctx context.Context, opts dto.PatchUpdateSongOpts) (dto.SongInfo, error) {
	filter := storageDto.UpdateSongInfoFilter{
		SongID:      opts.SongID,
		Title:       opts.Title,
		ArtistID:    opts.ArtistID,
		Lyrics:      opts.Lyrics,
		ReleaseDate: opts.ReleaseDate,
	}

	song, err := s.storage.UpdateSongInfo(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrSongNotFound) {
			return dto.SongInfo{}, fmt.Errorf("patch update song: %w", ErrSongNotFound)
		}
		return dto.SongInfo{}, fmt.Errorf("patch update song: %v", err)
	}

	return dto.SongInfo{
		SongID:      song.SongID,
		Title:       song.Title,
		Lyrics:      song.Lyrics,
		CoverURL:    song.CoverURL,
		ReleaseDate: song.ReleaseDate,
		Views:       song.Views,
		Artist: dto.Artist{
			ArtistID:  song.Artist.ArtistID,
			Name:      song.Artist.Name,
			Bio:       song.Artist.Bio,
			AvatarURL: song.Artist.AvatarURL,
		},
	}, nil
}

func (s *Storage) PatchUpdateCover(ctx context.Context, opts dto.PatchUpdateCoverOpts) error {
	filter := storageDto.UpdateSongCoverFilter{
		SongID:   opts.SongID,
		CoverURL: opts.CoverURL,
	}
	err := s.storage.UpdateSongCover(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrSongNotFound) {
			return fmt.Errorf("patch update cover: %w", ErrSongNotFound)
		}
		return fmt.Errorf("patch update cover: %v", err)
	}

	return nil
}
