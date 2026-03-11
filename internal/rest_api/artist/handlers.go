package artist

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/K1tten2005/lyryx-backend/internal/model/roles"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/middlewares"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/artist"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/artist/dto"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type artistUsecase interface {
	GetArtistByID(ctx context.Context, artistID int) (dto.Artist, error)
	PostArtist(ctx context.Context, opts dto.PostArtistOpts) (dto.Artist, error)
	PatchUpdateArtist(ctx context.Context, opts dto.PatchUpdateArtistOpts) (dto.Artist, error)
	PatchUpdateAvatar(ctx context.Context, opts dto.UploadAvatarOpts) (string, error)
}

type claimsGetter interface {
	GetClaims(c echo.Context) (*auth.JwtCustomClaims, error)
}

type Handlers struct {
	artistUsecase artistUsecase
	claimsGetter  claimsGetter

	logger *logrus.Logger
}

func NewArtistHandlers(
	artistUsecase artistUsecase,
	claimsGetter claimsGetter,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		artistUsecase: artistUsecase,
		claimsGetter:  claimsGetter,
		logger:        logger,
	}
}

func (h *Handlers) RegisterHandlers(e *echo.Echo, authMiddleware echo.MiddlewareFunc, checkRoleMiddleware *middlewares.RolesCheckerMiddleware) {
	public := e.Group("")
	public.GET("/v1/artist/:id", h.GetArtistByID)

	private := e.Group("")
	private.Use(authMiddleware)
	private.POST("/v1/artist", h.PostArtist, checkRoleMiddleware.CheckRole(roles.RoleModerator))
	private.PATCH("/v1/artist/:id", h.PatchUpdateArtist, checkRoleMiddleware.CheckRole(roles.RoleModerator))
	private.PATCH("/v1/artist/:id/avatar", h.PatchUpdateAvatar, checkRoleMiddleware.CheckRole(roles.RoleModerator))
}

// GetArtistByID godoc
// @Summary      Получение данных артиста по его id.
// @Description  Возвращает полную информацию о профиле артиста по его id.
// @Tags         artist
// @Produce      json
// @Param        id   path int      true  "Artist ID"
// @Success      200    {object} GetArtistByIDOut       "Успешный ответ с профилем артиста"
// @Failure      400    {object} echo.HTTPError      "Некорректный id артиста"
// @Failure      404    {object} echo.HTTPError      "Артист не найден"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/artist/{id} [get]
func (h *Handlers) GetArtistByID(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(GetArtistByIDIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	user, err := h.artistUsecase.GetArtistByID(ctx, req.ArtistID)
	if err != nil {
		h.logger.WithError(err).Warning(err.Error())
		if errors.Is(err, usecase.ErrArtistNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "artist not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	out := getArtistByIDToOut(user)

	return c.JSON(http.StatusOK, out)
}

func getArtistByIDToOut(artist dto.Artist) GetArtistByIDOut {
	return GetArtistByIDOut{
		ArtistID:  artist.ArtistID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}
}

// PostArtist godoc
// @Summary      Создание артиста.
// @Description  Создает нового артиста по имени и био. Доступно только модератору.
// @Tags         artist
// @Accept       json
// @Produce      json
// @Param        request body PostArtistIn true "Параметры артиста"
// @Success      201    {object} PostArtistOut      "Артист успешно создан"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      403    {object} echo.HTTPError      "Недостаточно прав"
// @Failure      409    {object} echo.HTTPError      "Имя артиста уже занято"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/artist [post]
func (h *Handlers) PostArtist(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(PostArtistIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts := postArtistToOpts(req)

	artist, err := h.artistUsecase.PostArtist(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning(err.Error())
		if errors.Is(err, usecase.ErrNameTaken) {
			return echo.NewHTTPError(http.StatusConflict, echo.Map{"error": "artist name is already taken"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "Invalid input"})
	}

	out := postArtistToOut(artist)

	return c.JSON(http.StatusCreated, out)
}

func postArtistToOpts(req *PostArtistIn) dto.PostArtistOpts {
	return dto.PostArtistOpts{
		Name: strings.TrimSpace(req.Name),
		Bio:  strings.TrimSpace(req.Bio),
	}
}

func postArtistToOut(artist dto.Artist) PostArtistOut {
	return PostArtistOut{
		ArtistID:  artist.ArtistID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}
}

// PatchUpdateArtist godoc
// @Summary      Частичное обновление профиля артиста
// @Description  Обновляет name, bio артиста по его id. Достаточно передать только нужные поля.
// @Tags         artist
// @Accept       json
// @Produce      json
// @Param        request body PatchUpdateArtistIn true "Параметры обновления профиля артиста"
// @Success      200    {object} PatchUpdateArtistOut   "Профиль артиста обновлен"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Артист не найден"
// @Failure      409    {object} echo.HTTPError      "Артист с таким именем уже существует"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/artist/{id} [patch]
func (h *Handlers) PatchUpdateArtist(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(PatchUpdateArtistIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts, err := patchUpdateArtistToOpts(req)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	artist, err := h.artistUsecase.PatchUpdateArtist(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("patch update artist failed")
		if errors.Is(err, usecase.ErrNameTaken) {
			return echo.NewHTTPError(http.StatusConflict, echo.Map{"error": "this username is already taken"})
		}
		if errors.Is(err, usecase.ErrArtistNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "artist not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	out := patchUpdateArtistToOut(artist)

	return c.JSON(http.StatusOK, out)
}

func patchUpdateArtistToOpts(req *PatchUpdateArtistIn) (dto.PatchUpdateArtistOpts, error) {
	if req.Name == nil && req.Bio == nil {
		return dto.PatchUpdateArtistOpts{}, errors.New("at least one field must be provided")
	}

	var name *string
	if req.Name != nil {
		normalizedName := strings.TrimSpace(*req.Name)
		name = &normalizedName
	}

	var bio *string
	if req.Bio != nil {
		normalizedBio := strings.TrimSpace(*req.Bio)
		bio = &normalizedBio
	}

	return dto.PatchUpdateArtistOpts{
		ArtistID: req.ArtistID,
		Name:     name,
		Bio:      bio,
	}, nil
}

func patchUpdateArtistToOut(artist dto.Artist) PatchUpdateArtistOut {
	return PatchUpdateArtistOut{
		ArtistID:  artist.ArtistID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}
}

// PatchUpdateAvatar godoc
// @Summary      Обновление аватарки артиста
// @Description  Принимает avatar в multipart/form-data, валидирует изображение, загружает в MinIO и обновляет ссылку на аватар артиста.
// @Tags         artist
// @Accept       mpfd
// @Produce      json
// @Param        avatar formData file true "Файл аватарки (png/jpeg)"
// @Success      200    {object} PatchUpdateAvatarOut "Аватар артиста обновлен"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Артист не найден"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /v1/artist/{id}/avatar [patch]
func (h *Handlers) PatchUpdateAvatar(c echo.Context) error {
	ctx := c.Request().Context()

	avatarFile, err := c.FormFile("avatar")
	if err != nil {
		h.logger.WithError(err).Warning("get avatar file failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "avatar file is required"})
	}

	artistIDstr := c.Param("id")
	if artistIDstr == "" {
		h.logger.WithError(err).Warning("get artist id failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "artist id is required"})
	}

	artistID, err := strconv.Atoi(artistIDstr)
	if err != nil {
		h.logger.WithError(err).Warning("get artist id failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid artist id"})
	}

	avatarUrl, err := h.artistUsecase.PatchUpdateAvatar(ctx, dto.UploadAvatarOpts{
		ArtistID:   artistID,
		AvatarFile: avatarFile,
	})
	if err != nil {
		h.logger.WithError(err).Warning("patch update avatar failed")
		if errors.Is(err, usecase.ErrAvatarTooLarge) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		if errors.Is(err, usecase.ErrInvalidAvatarType) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		if errors.Is(err, usecase.ErrArtistNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "artist not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, PatchUpdateAvatarOut{
		AvatarURL: avatarUrl,
	})
}
