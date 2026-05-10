package song

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/K1tten2005/lyryx-backend/internal/model/roles"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/middlewares"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/song"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/song/dto"
	timeHelper "github.com/K1tten2005/lyryx-backend/internal/usecases/utils/time_helper"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	ErrSongNotFound = errors.New("song not found")
)

type songUsecase interface {
	GetSongByID(ctx context.Context, songID int) (dto.SongInfo, error)
	PostSong(ctx context.Context, opts dto.PostSongOpts) (dto.SongInfo, error)
	PatchUpdateSong(ctx context.Context, opts dto.PatchUpdateSongOpts) (dto.SongInfo, error)
	PatchUpdateCover(ctx context.Context, opts dto.UploadCoverOpts) (string, error)
	GetAiTranslation(ctx context.Context, opts dto.GetAiTranslationOpts) (dto.AiTranslationResp, error)
}

type claimsGetter interface {
	GetClaims(c echo.Context) (*auth.JwtCustomClaims, error)
}

type Handlers struct {
	songUsecase  songUsecase
	claimsGetter claimsGetter

	logger *logrus.Logger
}

func NewSongHandlers(
	songUsecase songUsecase,
	claimsGetter claimsGetter,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		songUsecase:  songUsecase,
		claimsGetter: claimsGetter,
		logger:       logger,
	}
}

func (h *Handlers) RegisterHandlers(e *echo.Echo, authMiddleware echo.MiddlewareFunc, checkRoleMiddleware *middlewares.RolesCheckerMiddleware) {
	public := e.Group("")
	public.GET("/v1/song/:id", h.GetSongByID)

	private := e.Group("")
	private.Use(authMiddleware)
	private.POST("/v1/song", h.PostSong, checkRoleMiddleware.CheckRole(roles.RoleModerator))
	private.PATCH("/v1/song/:id", h.PatchUpdateSong, checkRoleMiddleware.CheckRole(roles.RoleModerator))
	private.PATCH("/v1/song/:id/cover", h.PatchUpdateCover, checkRoleMiddleware.CheckRole(roles.RoleModerator))
	private.GET("/v1/song/:id/ai-translation", h.GetAiTranslation)
}

// GetSongByID godoc
// @Summary      Получение данных песни по ее id.
// @Description  Возвращает полную информацию о песне по ее id.
// @Tags         artist
// @Produce      json
// @Param        id   path int      true  "Song ID"
// @Success      200    {object} GetSongByIDOut       "Успешный ответ с информацией о песне"
// @Failure      400    {object} echo.HTTPError      "Некорректный id песни"
// @Failure      404    {object} echo.HTTPError      "Песня не найдена"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/song/{id} [get]
func (h *Handlers) GetSongByID(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(GetSongByIDIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	user, err := h.songUsecase.GetSongByID(ctx, req.SongID)
	if err != nil {
		h.logger.WithError(err).Warning(err.Error())
		if errors.Is(err, usecase.ErrSongNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "artist not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	out := getSongByIDToOut(user)

	return c.JSON(http.StatusOK, out)
}

func getSongByIDToOut(song dto.SongInfo) GetSongByIDOut {
	return GetSongByIDOut{
		SongID:      song.SongID,
		Title:       song.Title,
		Lyrics:      song.Lyrics,
		CoverURL:    song.CoverURL,
		ReleaseDate: song.ReleaseDate,
		Views:       song.Views,
		Artist: Artist{
			ArtistID:  song.Artist.ArtistID,
			Name:      song.Artist.Name,
			Bio:       song.Artist.Bio,
			AvatarURL: song.Artist.AvatarURL,
		},
	}
}

// PostSong godoc
// @Summary      Создание песни.
// @Description  Создает новой песни по заданным параметрам. Доступно только модератору.
// @Tags         song
// @Accept       json
// @Produce      json
// @Param        request body PostSongIn true "Параметры песни"
// @Success      201    {object} PostSongOut      "Песня успешно создана"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      403    {object} echo.HTTPError      "Недостаточно прав"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/song [post]
func (h *Handlers) PostSong(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(PostSongIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts, err := postSongToOpts(req)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	artist, err := h.songUsecase.PostSong(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "Invalid input"})
	}

	out := postSongToOut(artist)

	return c.JSON(http.StatusCreated, out)
}

func postSongToOpts(req *PostSongIn) (dto.PostSongOpts, error) {
	releaseDate, err := timeHelper.StringToTime(req.ReleaseDate)
	if err != nil {
		return dto.PostSongOpts{}, fmt.Errorf("invalid release date: %s", req.ReleaseDate)
	}
	return dto.PostSongOpts{
		Title:       strings.TrimSpace(req.Title),
		Lyrics:      strings.TrimSpace(req.Lyrics),
		CoverURL:    strings.TrimSpace(req.CoverURL),
		ReleaseDate: releaseDate,
		ArtistID:    req.ArtistID,
	}, nil
}

func postSongToOut(song dto.SongInfo) PostSongOut {
	return PostSongOut{
		SongID:      song.SongID,
		Title:       song.Title,
		Lyrics:      song.Lyrics,
		CoverURL:    song.CoverURL,
		ReleaseDate: song.ReleaseDate,
		Views:       song.Views,
		Artist: Artist{
			ArtistID:  song.Artist.ArtistID,
			Name:      song.Artist.Name,
			Bio:       song.Artist.Bio,
			AvatarURL: song.Artist.AvatarURL,
		},
	}
}

// PatchUpdateSong godoc
// @Summary      Частичное обновление информации о песне
// @Description  Обновляет title, lyrics, artist_id, release_date песни по ее id. Достаточно передать только нужные поля.
// @Tags         song
// @Accept       json
// @Produce      json
// @Param        request body PatchUpdateSongIn true "Параметры обновления информации о песне"
// @Success      200    {object} PatchUpdateSongOut   "Страница песни обновлена"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Песня не найдена"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/song/{id} [patch]
func (h *Handlers) PatchUpdateSong(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(PatchUpdateSongIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts, err := patchUpdateSongToOpts(req)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	artist, err := h.songUsecase.PatchUpdateSong(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("patch update song failed")
		if errors.Is(err, usecase.ErrSongNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": usecase.ErrSongNotFound.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	out := patchUpdateSongToOut(artist)

	return c.JSON(http.StatusOK, out)
}

func patchUpdateSongToOpts(req *PatchUpdateSongIn) (dto.PatchUpdateSongOpts, error) {
	var title *string
	if req.Title != nil {
		normalizedTitle := strings.TrimSpace(*req.Title)
		title = &normalizedTitle
	}

	var lyrics *string
	if req.Lyrics != nil {
		normalizedLyrics := strings.TrimSpace(*req.Lyrics)
		lyrics = &normalizedLyrics
	}

	var releaseDatePtr *time.Time
	if req.ReleaseDate != nil {
		releaseDate, err := timeHelper.StringToTime(*req.ReleaseDate)
		if err != nil {
			return dto.PatchUpdateSongOpts{}, fmt.Errorf("invalid release date")
		}
		releaseDatePtr = &releaseDate
	}

	return dto.PatchUpdateSongOpts{
		SongID:      req.SongID,
		ArtistID:    req.ArtistID,
		Title:       title,
		Lyrics:      lyrics,
		ReleaseDate: releaseDatePtr,
	}, nil
}

func patchUpdateSongToOut(song dto.SongInfo) PatchUpdateSongOut {
	return PatchUpdateSongOut{
		SongID:      song.SongID,
		Title:       song.Title,
		Lyrics:      song.Lyrics,
		CoverURL:    song.CoverURL,
		ReleaseDate: song.ReleaseDate,
		Views:       song.Views,
		Artist: Artist{
			ArtistID:  song.Artist.ArtistID,
			Name:      song.Artist.Name,
			Bio:       song.Artist.Bio,
			AvatarURL: song.Artist.AvatarURL,
		},
	}
}

// PatchUpdateCover godoc
// @Summary      Обновление обложки песни
// @Description  Принимает cover в multipart/form-data, валидирует изображение, загружает в MinIO и обновляет ссылку на обложку песни.
// @Tags         song
// @Accept       mpfd
// @Produce      json
// @Param        avatar formData file true "Файл обложки (png/jpeg)"
// @Success      200    {object} PatchUpdateCoverOut "Обложка песни обновлена"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Песня не найдена"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/song/{id}/cover [patch]
func (h *Handlers) PatchUpdateCover(c echo.Context) error {
	ctx := c.Request().Context()

	coverFile, err := c.FormFile("cover")
	if err != nil {
		h.logger.WithError(err).Warning("get cover file failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "cover file is required"})
	}

	artistIDstr := c.Param("id")
	if artistIDstr == "" {
		h.logger.WithError(err).Warning("get song id failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "song id is required"})
	}

	songID, err := strconv.Atoi(artistIDstr)
	if err != nil {
		h.logger.WithError(err).Warning("get song id failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid song id"})
	}

	avatarUrl, err := h.songUsecase.PatchUpdateCover(ctx, dto.UploadCoverOpts{
		SongID:    songID,
		CoverFile: coverFile,
	})
	if err != nil {
		h.logger.WithError(err).Warning("patch update cover failed")
		if errors.Is(err, usecase.ErrCoverTooLarge) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		if errors.Is(err, usecase.ErrInvalidCoverType) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		if errors.Is(err, usecase.ErrSongNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "song not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, PatchUpdateCoverOut{
		CoverURL: avatarUrl,
	})
}

// GetAiTranslation godoc
// @Summary      Получение AI-перевода текста песни
// @Description  Генерирует перевод или объяснение текста песни с помощью искусственного интеллекта на указанный язык. Требуется аутентификация.
// @Tags         song
// @Produce      json
// @Param        id         path   int     true  "Song ID"
// @Param        language   query  string  true  "Код целевого языка (например: 'en', 'de', 'fr')"  minlength(2)  maxlength(5)
// @Success      200        {object} GetAiTranslationOut  "Успешный ответ с AI-переводом"
// @Failure      400        {object} echo.HTTPError      "Некорректные параметры запроса"
// @Failure      401        {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404        {object} echo.HTTPError      "Песня не найдена"
// @Failure      500        {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/song/{id}/ai-translation [get]
func (h *Handlers) GetAiTranslation(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(GetAiTranslationIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	opts := getAiTranslationToOpts(req)

	resp, err := h.songUsecase.GetAiTranslation(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("get ai annotation failed")
		if errors.Is(err, usecase.ErrSongNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrSongNotFound.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := GetAiTranslationOut{
		SongID:   req.SongID,
		Response: resp.Response,
	}

	return c.JSON(http.StatusOK, out)
}

func getAiTranslationToOpts(req *GetAiTranslationIn) dto.GetAiTranslationOpts {
	return dto.GetAiTranslationOpts{
		SongID:   req.SongID,
		Language: req.Language,
	}
}
