package auth

import (
	"context"
	"errors"
	"fmt"
	"lyryx-backend/internal/rest_api/utils/user_validation"
	"lyryx-backend/internal/usecases/auth/dto"
	usecase "lyryx-backend/internal/usecases/auth"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	accessTokenExp  = time.Hour * 24
	refreshTokenExp = time.Hour * 24 * 30
)

type authGetter interface {
	PostSignUp(ctx context.Context, opts dto.SignUpOpts) (dto.UserInfo, error)
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
	//public.POST("/v1/auth/sign-in", h.PostSignIn)

	private := e.Group("")
	private.Use(authMiddleware)
}

func generateTokens(userInfo *dto.UserInfo, jwtSecret []byte) (
	signedAccessToken string,
	signedRefreshToken string,
	err error,
) {
	accessClaims := &JwtCustomClaims{
		UserID: userInfo.UserID,
		Email:  userInfo.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExp)),
		},
	}

	refreshClaims := &JwtCustomClaims{
		Email: userInfo.Email,
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
			return echo.NewHTTPError(http.StatusConflict, "this email is already busy")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	signedAccessToken, signedRefreshToken, err := generateTokens(&userInfo, h.jwtSecret)
	if err != nil {
		h.logger.WithError(err).Warning("generate tokens failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
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

	return c.JSON(http.StatusOK, PostSignUpOut{
		AccessToken: signedAccessToken,
	})
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
