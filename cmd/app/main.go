package main

import (
	"context"

	"github.com/K1tten2005/lyryx-backend/internal/config"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	authHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	userHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/user"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/utils"
	authUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/auth"
	authStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/storage"
	authWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/wrappers"
	userUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/user"
	userStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/user/storage"
	userWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/user/wrappers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

// @title     lyryx API
// @version   1.0
// @host      localhost:8080
// @BasePath  /
func main() {
	// Logger.
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})

	// Config.
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Errorf("failed parse config: %v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %+v", r)
		}
	}()

	// Postgres.
	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Errorf("failed sqlx connect: %v", err)
		return
	}

	// Minio.
	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Errorf("failed create minio client: %v", err)
		return
	}

	// Auth middleware.
	authConfig := echojwt.Config{
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(authHandlersPkg.JwtCustomClaims)
		},
		SigningKey: []byte(cfg.JWTSecret),
	}

	authMiddleware := echojwt.WithConfig(authConfig)

	echoHandler := echo.New()

	// Validator.
	echoHandler.Validator = utils.NewHTTPRequestValidator()

	// Claims getter.
	claimsGetter := auth.ClaimsGetter{}

	authStorage := authStoragePkg.NewStorage(db)
	authWrappers := authWrappersPkg.NewStorage(authStorage)
	authUsecase := authUsecasePkg.NewUsecase(authWrappers, &authUsecasePkg.BcryptHasher{})
	authHandlers := authHandlersPkg.New(authUsecase, []byte(cfg.JWTSecret), logger)
	authHandlers.RegisterHandlers(echoHandler, authMiddleware)

	userStorage := userStoragePkg.NewStorage(db, logger)
	avatarStorage := userStoragePkg.NewMinIOAvatarStorage(minioClient, cfg.MinIOBucket, cfg.MinIOPublicBaseURL)
	if err := avatarStorage.EnsureBucketPublic(context.Background()); err != nil {
		log.Errorf("failed ensure minio bucket public policy: %v", err)
		return
	}
	userWrappers := userWrappersPkg.NewStorage(userStorage, avatarStorage)
	avatarWrapper := userWrappersPkg.NewAvatarGetter(avatarStorage)
	userUsecase := userUsecasePkg.NewUsecase(userWrappers, avatarWrapper, logger)
	userHandlers := userHandlersPkg.NewUserHandlers(userUsecase, claimsGetter, logger)
	userHandlers.RegisterHandlers(echoHandler, authMiddleware)

	echoHandler.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"*"}, // На этапе разработки можно всё
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	server := rest_api.NewServer(echoHandler)
	logger.Info("Trying to start server...")
	if err := server.Start(); err != nil {
		log.Errorf("server: %v", err)
	}
}
