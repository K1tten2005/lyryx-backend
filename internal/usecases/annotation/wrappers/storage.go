package wrappers

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/dto"
	storageDto "github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/storage"
)

var (
	ErrAnnotationsNotFound = errors.New("annotations not found")
	ErrForbidden           = errors.New("forbidden")
	ErrAnnotationNotFound  = errors.New("annotation not found")
	ErrSongNotFound        = errors.New("song not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrAnnotationOverlap   = errors.New("annotation overlaps with an existing one")
)

// Интерфейс реального хранилища (БД)
type storage interface {
	GetAnnotations(ctx context.Context, filter storageDto.GetAnnotationsFilter) ([]storageDto.AnnotationInfo, error)
	GetAnnotationByID(ctx context.Context, filter storageDto.GetAnnotationByIDFilter) (storageDto.AnnotationInfo, error)
	CreateAnnotation(ctx context.Context, filter storageDto.CreateAnnotationFilter) (storageDto.AnnotationInfo, error)
	UpdateAnnotation(ctx context.Context, filter storageDto.UpdateAnnotationFilter) (storageDto.AnnotationInfo, error)
	DeleteAnnotation(ctx context.Context, filter storageDto.DeleteAnnotationFilter) error
	VoteAnnotation(ctx context.Context, filter storageDto.VoteAnnotationFilter) (int, error)
	DeleteVote(ctx context.Context, filter storageDto.RemoveVoteFilter) error
	GetUserAnnotations(ctx context.Context, filter storageDto.GetUserAnnotationsFilter) ([]storageDto.AnnotationInfo, int, error)
}

// StorageWrapper адаптирует вызовы между usecase dto и storage dto
type StorageWrapper struct {
	storage storage
}

func NewStorage(storage storage) *StorageWrapper {
	return &StorageWrapper{
		storage: storage,
	}
}

// --- Helpers for mapping ---

func mapToDTOAnn(info storageDto.AnnotationInfo) dto.AnnotationInfo {
	return dto.AnnotationInfo{
		ID:         info.ID,
		Content:    info.Content,
		StartIndex: info.StartIndex,
		EndIndex:   info.EndIndex,
		Snippet:    info.Snippet,
		Rating:     info.Rating,
		CreatedAt:  info.CreatedAt,
		UpdatedAt:  info.UpdatedAt,
		MyVote:     info.MyVote,
		User: dto.UserInfo{
			UserID:          info.User.UserID,
			Username:        info.User.Username,
			AvatarURL:       info.User.AvatarURL,
			ReputationScore: info.User.ReputationScore,
		},
		Song: dto.SongInfo{
			ID:       info.Song.ID,
			Title:    info.Song.Title,
			CoverURL: info.Song.CoverURL,
			Artist: dto.ArtistInfo{
				ID:   info.Song.Artist.ID,
				Name: info.Song.Artist.Name,
			},
		},
	}
}

func mapToDTOAnnList(list []storageDto.AnnotationInfo) []dto.AnnotationInfo {
	res := make([]dto.AnnotationInfo, 0, len(list))
	for _, item := range list {
		res = append(res, mapToDTOAnn(item))
	}
	return res
}

// --- Implementation of UseCase Interface Methods ---

func (s *StorageWrapper) GetAnnotations(ctx context.Context, opts dto.GetAnnotationOpts) ([]dto.AnnotationInfo, error) {
	filter := storageDto.GetAnnotationsFilter{
		SongID: opts.SongID,
		UserID: opts.UserID,
	}

	anns, err := s.storage.GetAnnotations(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrAnnotationsNotFound) {
			return nil, fmt.Errorf("get annotations: %w", ErrAnnotationsNotFound)
		}
		if errors.Is(err, storageDto.ErrSongNotFound) {
			return nil, fmt.Errorf("get annotations: %w", ErrSongNotFound)
		}
		return nil, fmt.Errorf("get annotations: %w", err)
	}

	return mapToDTOAnnList(anns), nil
}

func (s *StorageWrapper) GetAnnotationByID(ctx context.Context, opts dto.GetAnnotationByIDOpts) (dto.AnnotationInfo, error) {
	filter := storageDto.GetAnnotationByIDFilter{
		AnnotationID: opts.AnnotationID,
		UserID:       opts.UserID,
	}

	info, err := s.storage.GetAnnotationByID(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrAnnotationNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("get annotation by id: %w", ErrAnnotationNotFound)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("get annotation by id: %w", err)
	}

	return mapToDTOAnn(info), nil
}

func (s *StorageWrapper) CreateAnnotation(ctx context.Context, opts dto.PostAnnotationOpts) (dto.AnnotationInfo, error) {
	filter := storageDto.CreateAnnotationFilter{
		AuthorID:   opts.UserID,
		SongID:     opts.SongID,
		Content:    opts.Content,
		StartIndex: opts.StartIndex,
		EndIndex:   opts.EndIndex,
	}

	info, err := s.storage.CreateAnnotation(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrSongNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("create annotation: %w", ErrSongNotFound)
		}
		if errors.Is(err, storageDto.ErrAnnotationOverlap) {
			return dto.AnnotationInfo{}, fmt.Errorf("create annotation: %w", ErrAnnotationOverlap)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("create annotation: %w", err)
	}

	return mapToDTOAnn(info), nil
}

func (s *StorageWrapper) UpdateAnnotation(ctx context.Context, opts dto.PatchUpdateAnnotationOpts) (dto.AnnotationInfo, error) {
	filter := storageDto.UpdateAnnotationFilter{
		AnnotationID: opts.AnnotationID,
		UserID:       opts.UserID,
		Content:      opts.Content,
	}

	info, err := s.storage.UpdateAnnotation(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrAnnotationNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("update annotation: %w", ErrAnnotationNotFound)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("update annotation: %w", err)
	}

	return mapToDTOAnn(info), nil
}

func (s *StorageWrapper) DeleteAnnotation(ctx context.Context, opts dto.DeleteAnnotationOpts) error {
	filter := storageDto.DeleteAnnotationFilter{
		AnnotationID: opts.AnnotationID,
		UserID:       opts.UserID,
		Role:         opts.Role,
	}

	err := s.storage.DeleteAnnotation(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrAnnotationNotFound) {
			return fmt.Errorf("delete annotation: %w", ErrAnnotationNotFound)
		}
		return fmt.Errorf("delete annotation: %w", err)
	}

	return nil
}

func (s *StorageWrapper) VoteAnnotation(ctx context.Context, opts dto.PostVoteAnnotationOpts) (int, error) {
	filter := storageDto.VoteAnnotationFilter{
		AnnotationID: opts.AnnotationID,
		UserID:       opts.UserID,
		Value:        opts.Value,
	}

	rating, err := s.storage.VoteAnnotation(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrAnnotationNotFound) {
			return 0, fmt.Errorf("vote annotation: %w", ErrAnnotationNotFound)
		}
		return 0, fmt.Errorf("vote annotation: %w", err)
	}

	return rating, nil
}

func (s *StorageWrapper) DeleteVote(ctx context.Context, opts dto.DeleteVoteOpts) error {
	filter := storageDto.RemoveVoteFilter{
		AnnotationID: opts.AnnotationID,
		UserID:       opts.UserID,
	}

	err := s.storage.DeleteVote(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrAnnotationNotFound) {
			return fmt.Errorf("remove vote: %w", ErrAnnotationNotFound)
		}
		return fmt.Errorf("remove vote: %w", err)
	}

	return nil
}

func (s *StorageWrapper) GetUserAnnotations(ctx context.Context, opts dto.GetUserAnnotationsOpts) ([]dto.AnnotationInfo, int, error) {
	filter := storageDto.GetUserAnnotationsFilter{
		UserID:        opts.UserID,
		CurrentUserID: opts.CurrentUserID,
		Limit:         opts.Limit,
		Offset:        opts.Offset,
	}

	anns, total, err := s.storage.GetUserAnnotations(ctx, filter)
	if err != nil {
		if errors.Is(err, storageDto.ErrUserNotFound) {
			return nil, 0, fmt.Errorf("get user annotations: %w", ErrUserNotFound)
		}
		return nil, 0, fmt.Errorf("get user annotations: %w", err)
	}

	return mapToDTOAnnList(anns), total, nil
}
