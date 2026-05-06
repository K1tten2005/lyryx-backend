// internal/usecases/search/usecase.go
package search

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/search/dto"
	"github.com/sirupsen/logrus"
)

var ErrResultNotFound = errors.New("no results")

type Storage interface {
	SearchSongs(ctx context.Context, opts dto.GetSearchOpts) ([]dto.SongInfo, error)
	SearchSongsByLyrics(ctx context.Context, opts dto.GetSearchOpts) ([]dto.SongInfo, error)
	SearchArtists(ctx context.Context, opts dto.GetSearchOpts) ([]dto.ArtistInfo, error)
	SearchUsers(ctx context.Context, opts dto.GetSearchOpts) ([]dto.UserInfo, error)
}

type Usecase struct {
	storage Storage
	logger  *logrus.Logger
}

func NewUsecase(storage Storage, logger *logrus.Logger) *Usecase {
	return &Usecase{storage: storage, logger: logger}
}

func (u *Usecase) GetSearch(ctx context.Context, opts dto.GetSearchOpts) (dto.SearchResult, error) {
	if opts.Limit == 0 {
		opts.Limit = 20
	}

	var (
		songs, lyrics []dto.SongInfo
		artists       []dto.ArtistInfo
		users         []dto.UserInfo
	)

	g, ctx := errgroup.WithContext(ctx)

	// Поиск песен по названию
	g.Go(func() error {
		var err error
		songs, err = u.storage.SearchSongs(ctx, opts)
		if err != nil {
			return fmt.Errorf("search songs: %w", err)
		}
		return nil
	})

	// Поиск песен по тексту
	g.Go(func() error {
		var err error
		lyrics, err = u.storage.SearchSongsByLyrics(ctx, opts)
		if err != nil {
			return fmt.Errorf("search lyrics: %w", err)
		}

		return nil
	})

	// Поиск артистов
	g.Go(func() error {
		var err error
		artists, err = u.storage.SearchArtists(ctx, opts)
		if err != nil {
			return fmt.Errorf("search artists: %w", err)
		}

		return nil
	})

	// Поиск пользователей
	g.Go(func() error {
		var err error
		users, err = u.storage.SearchUsers(ctx, opts)
		if err != nil {
			return fmt.Errorf("search users: %w", err)
		}
		
		return nil
	})

	if err := g.Wait(); err != nil {
		u.logger.WithError(err).Error("search failed")
		return dto.SearchResult{}, fmt.Errorf("search: %v", err)
	}

	// Если вообще ничего не найдено
	if len(songs)+len(lyrics)+len(artists)+len(users) == 0 {
		return dto.SearchResult{}, fmt.Errorf("search: %w", ErrResultNotFound)
	}

	return dto.SearchResult{
		Songs:              songs,
		LyricsMatchedSongs: lyrics,
		Artists:            artists,
		Users:              users,
	}, nil
}
