package song

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

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

type ollamaGetter interface {
	GetAitranslation(ctx context.Context, prompt string) (string, error)
}

type Usecase struct {
	storage           storage
	songCoverUploader songCoverUploader
	ollamaGetter      ollamaGetter

	logger *logrus.Logger
}

func NewUsecase(
	storage storage,
	songCoverUploader songCoverUploader,
	ollamaGetter ollamaGetter,

	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage:           storage,
		songCoverUploader: songCoverUploader,
		ollamaGetter:      ollamaGetter,
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
		SongID:   opts.SongID,
		CoverURL: coverUrl,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return "", ErrSongNotFound
		}
		return "", fmt.Errorf("patch update cover: %v", err)
	}

	return coverUrl, nil
}

func (u *Usecase) GetAiTranslation(ctx context.Context, opts dto.GetAiTranslationOpts) (dto.AiTranslationResp, error) {
	song, err := u.storage.GetSongByID(ctx, opts.SongID)
	if err != nil {
		if errors.Is(err, wrappers.ErrSongNotFound) {
			return dto.AiTranslationResp{}, fmt.Errorf("get song by id: %w", ErrSongNotFound)
		}
		return dto.AiTranslationResp{}, fmt.Errorf("get song by id: %v", err)
	}

	// Формируем промпт.
	prompt := u.buildAiPrompt(song, opts.Language)

	resp, err := u.ollamaGetter.GetAitranslation(ctx, prompt)
	if err != nil {
		return dto.AiTranslationResp{}, fmt.Errorf("get ai annotation: %v", err)
	}

	return dto.AiTranslationResp{
		Response: resp,
	}, nil
}

// buildAiPrompt формирует промпт для перевода текста песни
func (u *Usecase) buildAiPrompt(song dto.SongInfo, targetLanguage string) string {
	// Ограничиваем длину текста (модели имеют лимит контекста)
	lyrics := song.Lyrics
	const maxRunes = 3000 // ~750-1000 токенов для Qwen2.5:1.5B
	if utf8.RuneCountInString(lyrics) > maxRunes {
		lyrics = truncateLyrics(lyrics, maxRunes)
	}

	return fmt.Sprintf(`Ты — профессиональный переводчик песен с опытом работы в музыкальной индустрии.

**Задача:**
Переведи текст песни с оригинального языка на %s.

**Информация о песне:**
- Название: %s
- Артист: %s
- Оригинальный текст:
"""
%s
"""

**Требования к переводу:**
1. Сохраняй поэтическую структуру: строфы, ритм, рифму (где это возможно без искажения смысла)
2. Передавай эмоциональный тон и настроение оригинала
3. Культурные отсылки и идиомы:
   - Если есть точный аналог в языке перевода — используй его
   - Если нет — переводи описательно, сохраняя смысл
4. Не добавляй пояснений, комментариев или мета-текста в сам перевод
5. Если встречаются непереводимые слова (имена, названия) — оставляй их как есть
6. Избегай дословного перевода, если он ломает смысл или звучание

**Формат ответа:**
- Верни ТОЛЬКО переведённый текст, без пояснений, в таком же формате, как и в оригинале
- Названия куплетов и припевов в квадратных скобках не переводи
- Сохрани разбивку на строфы и строки как в оригинале
- Никакого Markdown, кавычек или дополнительного форматирования

Начинай перевод сразу с первой строки.`,
		targetLanguage,
		song.Title,
		song.Artist.Name,
		lyrics,
	)
}

// truncateLyrics обрезает текст, сохраняя начало и конец (если текст слишком длинный)
func truncateLyrics(lyrics string, maxRunes int) string {
	runes := []rune(lyrics)
	if len(runes) <= maxRunes {
		return lyrics
	}
	half := maxRunes / 2
	return string(runes[:half]) + "\n\n[... часть текста опущена ...]\n\n" + string(runes[len(runes)-half:])
}
