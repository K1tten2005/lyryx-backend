package wrappers

import (
	"context"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/search/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/search/storage"
)

type storage interface {
	SearchSongs(ctx context.Context, filter storageDto.GetSearchFilter) ([]storageDto.SongInfo, error)
	SearchSongsByLyrics(ctx context.Context, filter storageDto.GetSearchFilter) ([]storageDto.SongInfo, error)
	SearchArtists(ctx context.Context, filter storageDto.GetSearchFilter) ([]storageDto.ArtistInfo, error)
	SearchUsers(ctx context.Context, filter storageDto.GetSearchFilter) ([]storageDto.UserInfo, error)
}

type Storage struct {
	storage storage
}

func NewStorage(storage storage) *Storage {
	return &Storage{
		storage: storage,
	}
}

func (s *Storage) SearchSongs(ctx context.Context, opts dto.GetSearchOpts) ([]dto.SongInfo, error) {
	filter := storageDto.GetSearchFilter{
		Query: opts.Query,
		Limit: opts.Limit,
	}

	searchRes, err := s.storage.SearchSongs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get search: %v", err)
	}

	searchResUC := make([]dto.SongInfo, 0, len(searchRes))

	for _, song := range searchRes {
		searchResUC = append(searchResUC, dto.SongInfo{
			ID:            song.ID,
			Title:         song.Title,
			LyricsSnippet: song.LyricsSnippet,
			Artist: dto.ArtistInfo{
				ID:        song.Artist.ID,
				Name:      song.Artist.Name,
				AvatarURL: song.Artist.AvatarURL,
			},
			CoverURL: song.CoverURL,
		})
	}

	return searchResUC, nil
}

func (s *Storage) SearchSongsByLyrics(ctx context.Context, opts dto.GetSearchOpts) ([]dto.SongInfo, error) {
	filter := storageDto.GetSearchFilter{
		Query: opts.Query,
		Limit: opts.Limit,
	}

	searchRes, err := s.storage.SearchSongsByLyrics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get search: %v", err)
	}

	searchResUC := make([]dto.SongInfo, 0, len(searchRes))

	for _, song := range searchRes {
		searchResUC = append(searchResUC, dto.SongInfo{
			ID:            song.ID,
			Title:         song.Title,
			LyricsSnippet: song.LyricsSnippet,
			Artist: dto.ArtistInfo{
				ID:        song.Artist.ID,
				Name:      song.Artist.Name,
				AvatarURL: song.Artist.AvatarURL,
			},
			CoverURL: song.CoverURL,
		})
	}

	return searchResUC, nil
}

func (s *Storage) SearchArtists(ctx context.Context, opts dto.GetSearchOpts) ([]dto.ArtistInfo, error) {
	filter := storageDto.GetSearchFilter{
		Query: opts.Query,
		Limit: opts.Limit,
	}

	searchRes, err := s.storage.SearchArtists(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get search: %v", err)
	}

	searchResUC := make([]dto.ArtistInfo, 0, len(searchRes))

	for _, artist := range searchRes {
		searchResUC = append(searchResUC, dto.ArtistInfo{
			ID:        artist.ID,
			Name:      artist.Name,
			AvatarURL: artist.AvatarURL,
		})
	}

	return searchResUC, nil
}

func (s *Storage) SearchUsers(ctx context.Context, opts dto.GetSearchOpts) ([]dto.UserInfo, error) {
	filter := storageDto.GetSearchFilter{
		Query: opts.Query,
		Limit: opts.Limit,
	}

	searchRes, err := s.storage.SearchUsers(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get search: %v", err)
	}

	searchResUC := make([]dto.UserInfo, 0, len(searchRes))

	for _, user := range searchRes {
		searchResUC = append(searchResUC, dto.UserInfo{
			UserID:          user.UserID,
			Username:        user.Username,
			AvatarURL:       user.AvatarURL,
			ReputationScore: user.ReputationScore,
		})
	}

	return searchResUC, nil
}
