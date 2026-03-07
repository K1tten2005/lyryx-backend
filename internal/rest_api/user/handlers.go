package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/user"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type userUsecase interface {
	GetUserByID(ctx context.Context, userID int) (dto.User, error)
}

type claimsGetter interface {
	GetClaims(c echo.Context) (*auth.JwtCustomClaims, error)
}

type Handlers struct {
	userUsecase  userUsecase
	claimsGetter claimsGetter

	logger *logrus.Logger
}

func NewUserHandlers(
	userUsecase userUsecase,
	claimsGetter claimsGetter,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		userUsecase:  userUsecase,
		claimsGetter: claimsGetter,
		logger:       logger,
	}
}

func (h *Handlers) RegisterHandlers(e *echo.Echo, authMiddleware echo.MiddlewareFunc) {
	public := e.Group("")
	public.GET("/v1/user/:id", h.GetUserByID)

	private := e.Group("")
	private.Use(authMiddleware)
	private.GET("/v1/user/me", h.GetUserMe)
}

// GetUserMe godoc
// @Summary      Получение данных своего профиля.
// @Description  Возвращает полную информацию о профиле текущего пользователя, идентифицированного по access_token.
// @Tags         user
// @Produce      json
// @Success      200    {object} GetUserMeOut       "Успешный ответ с профилем пользователя"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/user/me [get]
func (h *Handlers) GetUserMe(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	logger := h.logger.WithFields(logrus.Fields{
		"userID": claims.UserID,
	})

	user, err := h.userUsecase.GetUserByID(ctx, claims.UserID)
	if err != nil {
		logger.WithError(err).Warning(err.Error())
		if errors.Is(err, usecase.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	out := getUserMeToOut(user)

	return c.JSON(http.StatusOK, out)
}

func getUserMeToOut(user dto.User) GetUserMeOut {
	return GetUserMeOut{
		UserID:          user.UserID,
		Email:           user.Email,
		Username:        user.Username,
		ReputationScore: user.ReputationScore,
		Role:            user.Role,
	}
}

// GetUserByID godoc
// @Summary      Получение данных пользователя по его id.
// @Description  Возвращает полную информацию о профиле пользователя по его id.
// @Tags         user
// @Produce      json
// @Param        id   path int      true  "User ID"
// @Success      200    {object} GetUserByIDOut       "Успешный ответ с профилем пользователя"
// @Failure      400    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Пользователь не найден"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/user/{id} [get]
func (h *Handlers) GetUserByID(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(GetUserByIDIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	user, err := h.userUsecase.GetUserByID(ctx, req.UserID)
	if err != nil {
		h.logger.WithError(err).Warning(err.Error())
		if errors.Is(err, usecase.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	out := getUserByIDToOut(user)

	return c.JSON(http.StatusOK, out)
}

func getUserByIDToOut(user dto.User) GetUserByIDOut {
	return GetUserByIDOut{
		UserID:          user.UserID,
		Email:           user.Email,
		Username:        user.Username,
		ReputationScore: user.ReputationScore,
		Role:            user.Role,
	}
}
