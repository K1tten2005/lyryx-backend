package user

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/utils/user_validation"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/user"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type userUsecase interface {
	GetUserByID(ctx context.Context, userID int) (dto.User, error)
	PatchUpdateUser(ctx context.Context, opts dto.PatchUpdateUserOpts) error
	PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error
}

type claimsGetter interface {
	GetClaims(c echo.Context) (*auth.JwtCustomClaims, error)
}

type avatarStorage interface {
	UploadAvatar(ctx context.Context, userID int, avatarFile *multipart.FileHeader) (string, error)
}

type Handlers struct {
	userUsecase  userUsecase
	claimsGetter claimsGetter
	avatarStore  avatarStorage

	logger *logrus.Logger
}

func NewUserHandlers(
	userUsecase userUsecase,
	claimsGetter claimsGetter,
	avatarStore avatarStorage,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		userUsecase:  userUsecase,
		claimsGetter: claimsGetter,
		avatarStore:  avatarStore,
		logger:       logger,
	}
}

func (h *Handlers) RegisterHandlers(e *echo.Echo, authMiddleware echo.MiddlewareFunc) {
	public := e.Group("")
	public.GET("/v1/user/:id", h.GetUserByID)

	private := e.Group("")
	private.Use(authMiddleware)
	private.GET("/v1/user/me", h.GetUserMe)
	private.PATCH("/v1/user/me", h.PatchUpdateUser)
	private.PATCH("/v1/user/me/avatar", h.PatchUpdateAvatar)
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
		Bio:             user.Bio,
		AvatarURL:       user.AvatarURL,
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
		Bio:             user.Bio,
		AvatarURL:       user.AvatarURL,
		ReputationScore: user.ReputationScore,
		Role:            user.Role,
	}
}

// PatchUpdateUser godoc
// @Summary      Частичное обновление профиля пользователя
// @Description  Обновляет email, username, bio или password текущего пользователя. Достаточно передать только нужные поля.
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request body PatchUpdateUserIn true "Параметры обновления профиля"
// @Success      204
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Пользователь не найден"
// @Failure      409    {object} echo.HTTPError      "Email или username уже заняты"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/user/me [patch]
func (h *Handlers) PatchUpdateUser(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	req := new(PatchUpdateUserIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts, err := patchUpdateUserToOpts(claims.UserID, req)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	err = h.userUsecase.PatchUpdateUser(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("patch update user failed")
		if errors.Is(err, usecase.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, echo.Map{"error": "this email is already taken"})
		}
		if errors.Is(err, usecase.ErrUsernameTaken) {
			return echo.NewHTTPError(http.StatusConflict, echo.Map{"error": "this username is already taken"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	return c.NoContent(http.StatusNoContent)
}

func patchUpdateUserToOpts(userID int, req *PatchUpdateUserIn) (dto.PatchUpdateUserOpts, error) {
	if req.Email == nil && req.Username == nil && req.Bio == nil && req.Password == nil {
		return dto.PatchUpdateUserOpts{}, errors.New("at least one field must be provided")
	}

	var email *string
	if req.Email != nil {
		normalizedEmail := strings.ToLower(strings.TrimSpace(*req.Email))
		if err := user_validation.ValidateEmail(normalizedEmail); err != nil {
			return dto.PatchUpdateUserOpts{}, fmt.Errorf("email validation failed: %v", err)
		}
		email = &normalizedEmail
	}

	var username *string
	if req.Username != nil {
		normalizedUsername := strings.TrimSpace(*req.Username)
		if err := user_validation.ValidateUsername(normalizedUsername); err != nil {
			return dto.PatchUpdateUserOpts{}, fmt.Errorf("username validation failed: %v", err)
		}
		username = &normalizedUsername
	}

	var bio *string
	if req.Bio != nil {
		normalizedBio := strings.TrimSpace(*req.Bio)
		bio = &normalizedBio
	}

	var password *string
	if req.Password != nil {
		if err := user_validation.ValidatePassword(*req.Password); err != nil {
			return dto.PatchUpdateUserOpts{}, fmt.Errorf("password validation failed: %v", err)
		}
		password = req.Password
	}

	return dto.PatchUpdateUserOpts{
		UserID:   userID,
		Email:    email,
		Username: username,
		Bio:      bio,
		Password: password,
	}, nil
}

// PatchUpdateAvatar godoc
// @Summary      Обновление аватарки пользователя
// @Description  Принимает avatar в multipart/form-data, валидирует изображение, загружает в MinIO и обновляет ссылку на аватар текущего пользователя.
// @Tags         user
// @Accept       mpfd
// @Produce      json
// @Param        avatar formData file true "Файл аватарки (png/jpeg/gif)"
// @Success      204
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      401    {object} echo.HTTPError      "Пользователь не аутентифицирован"
// @Failure      404    {object} echo.HTTPError      "Пользователь не найден"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/user/me/avatar [patch]
func (h *Handlers) PatchUpdateAvatar(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	avatarFile, err := c.FormFile("avatar")
	if err != nil {
		h.logger.WithError(err).Warning("get avatar file failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "avatar file is required"})
	}

	avatarURL, err := h.avatarStore.UploadAvatar(ctx, claims.UserID, avatarFile)
	if err != nil {
		h.logger.WithError(err).Warning("upload avatar failed")
		if errors.Is(err, ErrAvatarTooLarge) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		if errors.Is(err, ErrInvalidAvatarType) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	err = h.userUsecase.PatchUpdateAvatar(ctx, dto.PatchUpdateAvatarOpts{
		UserID:    claims.UserID,
		AvatarURL: avatarURL,
	})
	if err != nil {
		h.logger.WithError(err).Warning("patch update avatar failed")
		if errors.Is(err, usecase.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	return c.NoContent(http.StatusNoContent)
}
