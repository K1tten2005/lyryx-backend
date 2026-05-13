package annotation

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/K1tten2005/lyryx-backend/internal/model/roles"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/dto"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/wrappers"
	songDto "github.com/K1tten2005/lyryx-backend/internal/usecases/song/dto"
	userDto "github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	"github.com/sirupsen/logrus"
)

var (
	ErrAnnotationsNotFound = errors.New("annotations not found")
	ErrForbidden           = errors.New("forbidden")
	ErrAnnotationNotFound  = errors.New("annotation not found")
	ErrSongNotFound        = errors.New("song not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidIndex        = errors.New("start_index or end_index is out of range")
	ErrAnnotationOverlap   = errors.New("annotation overlaps with an existing one")
)

type storage interface {
	GetAnnotations(ctx context.Context, opts dto.GetAnnotationOpts) ([]dto.AnnotationInfo, error)
	GetAnnotationByID(ctx context.Context, opts dto.GetAnnotationByIDOpts) (dto.AnnotationInfo, error)
	CreateAnnotation(ctx context.Context, opts dto.PostAnnotationOpts) (dto.AnnotationInfo, error)
	UpdateAnnotation(ctx context.Context, opts dto.PatchUpdateAnnotationOpts) (dto.AnnotationInfo, error)
	DeleteAnnotation(ctx context.Context, opts dto.DeleteAnnotationOpts) error
	VoteAnnotation(ctx context.Context, opts dto.PostVoteAnnotationOpts) (int, error)
	DeleteVote(ctx context.Context, opts dto.DeleteVoteOpts) error
	GetUserAnnotations(ctx context.Context, opts dto.GetUserAnnotationsOpts) ([]dto.AnnotationInfo, int, error)
}

type songGetter interface {
	GetSongByID(ctx context.Context, artistID int) (songDto.SongInfo, error)
}

type userGetter interface {
	GetUserByID(ctx context.Context, userID int) (userDto.User, error)
}

type ollamaGetter interface {
	GetAiAnnotation(ctx context.Context, prompt string) (string, error)
}

type Usecase struct {
	storage      storage
	songGetter   songGetter
	userGetter   userGetter
	ollamaGetter ollamaGetter
	logger       *logrus.Logger
}

func NewUsecase(
	storage storage,
	songGetter songGetter,
	userGetter userGetter,
	ollamaGetter ollamaGetter,
	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage:      storage,
		songGetter:   songGetter,
		userGetter:   userGetter,
		ollamaGetter: ollamaGetter,
		logger:       logger,
	}
}

func (u *Usecase) GetAnnotations(ctx context.Context, opts dto.GetAnnotationOpts) ([]dto.AnnotationInfo, error) {
	_, err := u.songGetter.GetSongByID(ctx, opts.SongID)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return nil, fmt.Errorf("get song by id: %w", ErrSongNotFound)
		}
		return nil, fmt.Errorf("get song by id: %v", err)
	}

	annotations, err := u.storage.GetAnnotations(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationsNotFound) {
			return nil, fmt.Errorf("get annotations: %w", ErrAnnotationsNotFound)
		}
		return nil, fmt.Errorf("get annotations: %v", err)
	}

	return annotations, nil
}

func (u *Usecase) GetAnnotationByID(ctx context.Context, opts dto.GetAnnotationByIDOpts) (dto.AnnotationInfo, error) {
	annotation, err := u.storage.GetAnnotationByID(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("get annotation by id: %w", ErrAnnotationNotFound)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("get annotation by id: %v", err)
	}

	return annotation, nil
}

func (u *Usecase) CreateAnnotation(ctx context.Context, opts dto.PostAnnotationOpts) (dto.AnnotationInfo, error) {
	song, err := u.songGetter.GetSongByID(ctx, opts.SongID)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("post annotation: %w", ErrSongNotFound)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("post annotation: %v", err)
	}

	lyricsCount := utf8.RuneCountInString(song.Lyrics)
	if opts.StartIndex > lyricsCount || opts.EndIndex > lyricsCount {
		return dto.AnnotationInfo{}, fmt.Errorf("post annotation: %w", ErrInvalidIndex)
	}

	annotation, err := u.storage.CreateAnnotation(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationOverlap) {
			return dto.AnnotationInfo{}, fmt.Errorf("post annotation: %w", ErrAnnotationOverlap)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("post annotation: %v", err)
	}

	return annotation, nil
}

func (u *Usecase) UpdateAnnotation(ctx context.Context, opts dto.PatchUpdateAnnotationOpts) (dto.AnnotationInfo, error) {
	annotation, err := u.storage.GetAnnotationByID(ctx, dto.GetAnnotationByIDOpts{
		AnnotationID: opts.AnnotationID,
		UserID:       &opts.UserID,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("patch update annotation: %w", ErrAnnotationNotFound)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("patch update annotation: %v", err)
	}

	if annotation.User.UserID != opts.UserID {
		return dto.AnnotationInfo{}, fmt.Errorf("patch update annotation: %w", ErrForbidden)
	}

	annotation, err = u.storage.UpdateAnnotation(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return dto.AnnotationInfo{}, fmt.Errorf("patch update annotation: %w", ErrAnnotationNotFound)
		}
		return dto.AnnotationInfo{}, fmt.Errorf("patch update annotation: %v", err)
	}

	return annotation, nil
}

func (u *Usecase) DeleteAnnotation(ctx context.Context, opts dto.DeleteAnnotationOpts) error {
	annotation, err := u.storage.GetAnnotationByID(ctx, dto.GetAnnotationByIDOpts{
		AnnotationID: opts.AnnotationID,
		UserID:       &opts.UserID,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return fmt.Errorf("delete annotation: %w", ErrAnnotationNotFound)
		}
		return fmt.Errorf("delete annotation: %v", err)
	}

	// Проверяем, принадлежит ли удаляемая аннотация пользователю, или роль пользователя - модератор.
	isAuthor := annotation.User.UserID == opts.UserID
	isModerator := opts.Role == string(roles.RoleModerator)

	if !isAuthor && !isModerator {
		return fmt.Errorf("delete annotation: %w", ErrForbidden)
	}

	err = u.storage.DeleteAnnotation(ctx, opts)
	if err != nil {
		return fmt.Errorf("delete annotation: %v", err)
	}

	return nil
}

func (u *Usecase) VoteAnnotation(ctx context.Context, opts dto.PostVoteAnnotationOpts) (int, error) {
	_, err := u.storage.GetAnnotationByID(ctx, dto.GetAnnotationByIDOpts{
		AnnotationID: opts.AnnotationID,
		UserID:       &opts.UserID,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return 0, fmt.Errorf("vote annotation: %w", ErrAnnotationNotFound)
		}
		return 0, fmt.Errorf("vote annotation: %v", err)
	}

	newRating, err := u.storage.VoteAnnotation(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return 0, fmt.Errorf("vote annotation: %w", ErrAnnotationNotFound)
		}
		return 0, fmt.Errorf("vote annotation: %v", err)
	}

	return newRating, nil
}

func (u *Usecase) DeleteVote(ctx context.Context, opts dto.DeleteVoteOpts) error {
	err := u.storage.DeleteVote(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrAnnotationNotFound) {
			return fmt.Errorf("delete vote: %w", ErrAnnotationNotFound)
		}
		return fmt.Errorf("delete vote: %v", err)
	}

	return nil
}

func (u *Usecase) GetUserAnnotations(ctx context.Context, opts dto.GetUserAnnotationsOpts) ([]dto.AnnotationInfo, int, error) {
	_, err := u.userGetter.GetUserByID(ctx, opts.UserID)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserNotFound) {
			return nil, 0, fmt.Errorf("get user annotations: %w", ErrUserNotFound)
		}
		return nil, 0, fmt.Errorf("get user annotations: %v", err)
	}
	annotations, total, err := u.storage.GetUserAnnotations(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserNotFound) {
			return nil, 0, fmt.Errorf("get user annotations: %w", ErrUserNotFound)
		}
		return nil, 0, fmt.Errorf("get user annotations: %v", err)
	}

	return annotations, total, nil
}

func (u *Usecase) GetAiAnnotation(ctx context.Context, opts dto.GetAiAnnotationOpts) (dto.AiAnnotationResp, error) {
	song, err := u.songGetter.GetSongByID(ctx, opts.SongID)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return dto.AiAnnotationResp{}, fmt.Errorf("get song by id: %w", ErrSongNotFound)
		}
		return dto.AiAnnotationResp{}, fmt.Errorf("get song by id: %v", err)
	}

	// Формируем промпт.
	selectedLyrics := song.Lyrics[*opts.StartIndex:*opts.EndIndex]
	prompt := u.buildAiPrompt(song, selectedLyrics, opts.Question)

	resp, err := u.ollamaGetter.GetAiAnnotation(ctx, prompt)
	if err != nil {
		return dto.AiAnnotationResp{}, fmt.Errorf("get ai annotation: %v", err)
	}

	return dto.AiAnnotationResp{
		Response: resp,
	}, nil
}

// buildAiPrompt формирует структурированный промпт для LLM
func (u *Usecase) buildAiPrompt(song songDto.SongInfo, selectedLyrics, userQuestion string) string {
	// Ограничиваем длину текста, чтобы не превысить контекстное окно модели
	lyrics := song.Lyrics
	const maxLyricsLen = 2000 // в рунах, не байтах!
	if utf8.RuneCountInString(lyrics) > maxLyricsLen {
		lyrics = truncateLyrics(lyrics, maxLyricsLen)
	}

	return fmt.Sprintf(`Ты — эксперт по анализу текстов песен. Дай развёрнутый, но лаконичный ответ на вопрос пользователя.

**Информация о песне:**
- Название: %s
- Артист: %s
- Текст: 
"""
%s
"""
- Выделенный текст, к которому относится вопрос:
"""
%s
"""
**Вопрос пользователя:**
"%s"

**Требования к ответу:**
- Отвечай на русском языке
- Будь точен, опирайся на текст песни
- Если вопрос не относится к тексту — так и скажи
- Не выдумывай факты
- Максимальная длина ответа: 300 слов`,
		song.Title,
		song.Artist.Name,
		lyrics,
		selectedLyrics,
		userQuestion,
	)
}

// truncateLyrics обрезает текст, сохраняя начало и конец
func truncateLyrics(lyrics string, maxRunes int) string {
	runes := []rune(lyrics)
	if len(runes) <= maxRunes {
		return lyrics
	}
	half := maxRunes / 2
	return string(runes[:half]) + "\n\n[...текст сокращён...]\n\n" + string(runes[len(runes)-half:])
}
