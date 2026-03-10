package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/K1tten2005/lyryx-backend/internal/rest_api/utils/user_validation"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/auth"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/auth/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	accessTokenExp  = time.Minute * 15
	refreshTokenExp = time.Hour * 24 * 30
)

type authGetter interface {
	PostSignUp(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error)
	PostSignIn(ctx context.Context, opts dto.SignInOpts) (dto.UserInfo, error)
	SetNewRefreshToken(ctx context.Context, opts dto.SetNewRefreshTokenOpts) error
	PostSignOut(ctx context.Context, opts dto.SignOutOpts) error
	GetUserByRefreshToken(ctx context.Context, refreshToken string) (dto.UserInfo, error)
}

type ClaimsGetter struct{}

func (cg ClaimsGetter) GetClaims(c echo.Context) (*JwtCustomClaims, error) {
	return GetClaims(c)
}

// GetClaims достает claims из контекста.
func GetClaims(c echo.Context) (*JwtCustomClaims, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
	}

	return claims, nil
}

type Handlers struct {
	authGetter authGetter
	jwtSecret  []byte
	logger     *logrus.Logger
}

func New(
	authGetter authGetter,
	jwtSecret []byte,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		authGetter: authGetter,
		jwtSecret:  jwtSecret,
		logger:     logger,
	}
}

func (h *Handlers) RegisterHandlers(e *echo.Echo, authMiddleware echo.MiddlewareFunc) {
	public := e.Group("")
	public.POST("/v1/auth/sign-up", h.PostSignUp)
	public.POST("/v1/auth/sign-in", h.PostSignIn)
	public.POST("/v1/auth/refresh", h.PostRefreshToken)

	private := e.Group("")
	private.Use(authMiddleware)
	private.POST("/v1/auth/sign-out", h.PostSignOut)
}

func generateTokens(userInfo *dto.UserInfo, jwtSecret []byte) (
	signedAccessToken string,
	signedRefreshToken string,
	err error,
) {
	accessClaims := &JwtCustomClaims{
		UserID: userInfo.UserID,
		Email:  userInfo.Email,
		Role:   userInfo.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExp)),
		},
	}

	refreshClaims := &JwtCustomClaims{
		UserID: userInfo.UserID,
		Email:  userInfo.Email,
		Role:   userInfo.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExp)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	signedAccessToken, err = accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign jwt: %v", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	signedRefreshToken, err = refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign jwt: %v", err)
	}

	return signedAccessToken, signedRefreshToken, nil
}

// PostSignUp godoc
// @Summary       Регистрация пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param 	     request body   PostSignUpIn   true "Параметры запроса."
// @Success      200    {object} PostSignUpOut       "Успешный ответ с access_token"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      404    {object} echo.HTTPError      "Информация не найдена"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Failure      409    {object} echo.HTTPError      "Пользователь уже зарегистрирован"
// @Router       /v1/auth/sign-up [post]
func (h *Handlers) PostSignUp(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(PostSignUpIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts, err := signUpInToOpts(req)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// 1. Регистрация.
	userInfo, err := h.authGetter.PostSignUp(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("sign up failed")
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, echo.Map{"error": "this email is already taken"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	signedAccessToken, signedRefreshToken, err := generateTokens(&userInfo, h.jwtSecret)
	if err != nil {
		h.logger.WithError(err).Warning("generate tokens failed")
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	err = h.authGetter.SetNewRefreshToken(ctx, setNewRefreshTokenToOpts(userInfo.Email, signedRefreshToken))
	if err != nil {
		h.logger.WithError(err).Warning("set new refresh_token failed")
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	// 3. Установка refresh токена в куки.
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    signedRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Установите в true для HTTPS.
		MaxAge:   int(refreshTokenExp.Seconds()),
	})

	out := signUpOptsToOut(userInfo, signedAccessToken)

	return c.JSON(http.StatusCreated, out)
}

func signUpInToOpts(req *PostSignUpIn) (dto.SignUpOpts, error) {
	email := strings.ToLower(req.Email)
	if err := user_validation.ValidateEmail(email); err != nil {
		return dto.SignUpOpts{}, fmt.Errorf("email validation failed: %v", err)
	}

	if err := user_validation.ValidatePassword(req.Password); err != nil {
		return dto.SignUpOpts{}, fmt.Errorf("password validation failed: %v", err)
	}

	username := strings.TrimSpace(req.Username)
	if err := user_validation.ValidateUsername(username); err != nil {
		return dto.SignUpOpts{}, fmt.Errorf("username validation failed: %v", err)
	}

	return dto.SignUpOpts{
		Username: username,
		Email:    email,
		Password: req.Password,
	}, nil
}

func signUpOptsToOut(opts dto.UserInfo, accessToken string) PostSignUpOut {
	return PostSignUpOut{
		User: UserInfo{
			UserID:          opts.UserID,
			Email:           opts.Email,
			Username:        opts.Username,
			ReputationScore: opts.ReputationScore,
			Role:            opts.Role,
		},
		AccessToken: accessToken,
	}
}

// PostSignIn godoc
// @Summary       Авторизация пользователя
// @Tags         auth
// @Produce      json
// @Accept       json
// @Param 	     request body   PostSignInIn   true "Параметры запроса."
// @Success      200    {object} PostSignInOut       "Успешный ответ с access_token"
// @Failure      400    {object} echo.HTTPError      "Некорректный запрос"
// @Failure      404    {object} echo.HTTPError      "Информация не найдена"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/auth/sign-in [post]
func (h *Handlers) PostSignIn(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(PostSignInIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	opts, err := signInInToOpts(req)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// 1. Проверка учетных данных.
	userInfo, err := h.authGetter.PostSignIn(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("verify login failed")
		if errors.Is(err, usecase.ErrUserDoesntExist) || errors.Is(err, usecase.ErrInvalidPassword) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": "invalid email or password"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	// Создаем токены.
	signedAccessToken, signedRefreshToken, err := generateTokens(&userInfo, h.jwtSecret)
	if err != nil {
		h.logger.WithError(err).Warning("generate tokens failed")
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	// RefreshToken сохраняем в бд.
	err = h.authGetter.SetNewRefreshToken(ctx, setNewRefreshTokenToOpts(opts.Email, signedRefreshToken))
	if err != nil {
		h.logger.WithError(err).Warning("set new refresh_token failed")
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	// 3. Установка refresh токена в куки.
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    signedRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Установите в true для HTTPS.
		MaxAge:   int(refreshTokenExp.Seconds()),
	})

	out := signInOptsToOut(userInfo, signedAccessToken)

	return c.JSON(http.StatusOK, out)
}

func signInInToOpts(req *PostSignInIn) (dto.SignInOpts, error) {
	email := strings.ToLower(req.Email)
	if err := user_validation.ValidateEmail(email); err != nil {
		return dto.SignInOpts{}, fmt.Errorf("email validation failed: %v", err)
	}

	if err := user_validation.ValidatePassword(req.Password); err != nil {
		return dto.SignInOpts{}, fmt.Errorf("password validation failed: %v", err)
	}

	return dto.SignInOpts{
		Email:    email,
		Password: req.Password,
	}, nil
}

func signInOptsToOut(opts dto.UserInfo, accessToken string) PostSignInOut {
	return PostSignInOut{
		User: UserInfo{
			UserID:          opts.UserID,
			Email:           opts.Email,
			Username:        opts.Username,
			ReputationScore: opts.ReputationScore,
			Role:            opts.Role,
		},
		AccessToken: accessToken,
	}
}

func setNewRefreshTokenToOpts(email, refreshToken string) dto.SetNewRefreshTokenOpts {
	return dto.SetNewRefreshTokenOpts{
		Email:        email,
		RefreshToken: refreshToken,
	}
}

// PostSignOut godoc
// @Summary      Выход из аккаунта
// @Tags         auth
// @Produce      json
// @Success      200
// @Failure      401    {object} echo.HTTPError      "Пользователь не авторизован"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/auth/sign-out [post]
func (h *Handlers) PostSignOut(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := GetClaims(c)
	if err != nil {
		return err
	}

	err = h.authGetter.PostSignOut(ctx, dto.SignOutOpts{
		Email: claims.Email,
	})
	if err != nil {
		h.logger.WithError(err).Warning("sign out failed")
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	return c.NoContent(http.StatusOK)
}

// PostRefreshToken godoc
// @Summary      Проверка refresh токена
// @Tags         auth
// @Produce      json
// @Success      200    {object} PostRefreshTokenOut      "Успешное обновление авторизованности"
// @Failure      401    {object} echo.HTTPError      "Пользователь не авторизован"
// @Failure      500    {object} echo.HTTPError      "Внутренняя ошибка сервера"
// @Router       /v1/auth/refresh [post]
func (h *Handlers) PostRefreshToken(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. Получение refresh токена из куки.
	refreshTokenCookie, err := c.Cookie("refresh_token")
	if err != nil {
		h.logger.WithError(err).Warning("no cookie")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	refreshToken := refreshTokenCookie.Value

	// 2. Парсинг токена.
	token, err := jwt.ParseWithClaims(refreshToken, &JwtCustomClaims{}, func(_ *jwt.Token) (any, error) {
		return h.jwtSecret, nil
	})

	// 3. Проверка токена на валидность.
	if err != nil || token == nil || !token.Valid {
		h.logger.Warning("invalid token")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	// 4. Проверка что есть юзер с таким токеном.
	userInfo, err := h.authGetter.GetUserByRefreshToken(ctx, refreshToken)
	if err != nil {
		h.logger.WithError(err).Warning("failed to get user by refresh token")
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	// 5. Создание новых токенов.
	signedAccessToken, signedRefreshToken, err := generateTokens(&userInfo, h.jwtSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	// RefreshToken сохраняем в бд.
	err = h.authGetter.SetNewRefreshToken(ctx, setNewRefreshTokenToOpts(userInfo.Email, signedRefreshToken))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    signedRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Установите в true для HTTPS.
		MaxAge:   int(refreshTokenExp.Seconds()),
	})

	return c.JSON(http.StatusOK, PostRefreshTokenOut{
		AccessToken: signedAccessToken,
	})
}
