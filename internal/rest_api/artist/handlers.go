package artist

import (
	"context"
	"errors"
	"net/http"

	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/middlewares"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/artist/dto"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/artist"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type artistUsecase interface {
	GetArtistByID(ctx context.Context, artistID int) (dto.Artist, error)
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
	//private.POST("/v1/artist/", h.PostArtist, checkRoleMiddleware.CheckRole(roles.RoleModerator))
}

// GetArtistByID godoc
// @Summary      Получение данных артиста по его id.
// @Description  Возвращает полную информацию о профиле артиста по его id.
// @Tags         artist
// @Produce      json
// @Param        id   path int      true  "Artist ID"
// @Success      200    {object} GetArtistByIDOut       "Успешный ответ с профилем артиста"
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
