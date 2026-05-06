package main

import (
	"context"

	"github.com/K1tten2005/lyryx-backend/internal/config"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/middlewares"

	annotationHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/annotation"
	artistHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/artist"
	authHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	songHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/song"
	userHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/user"
	searchHandlersPkg "github.com/K1tten2005/lyryx-backend/internal/rest_api/search"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/utils"
	annotationUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/annotation"
	annotationStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/storage"
	annotationWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/wrappers"
	artistUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/artist"
	artistStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/artist/storage"
	artistWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/artist/wrappers"
	searchUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/search"
	searchStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/search/storage"
	searchWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/search/wrappers"
	authUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/auth"
	authStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/storage"
	authWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/wrappers"
	songUsecasePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/song"
	songStoragePkg "github.com/K1tten2005/lyryx-backend/internal/usecases/song/storage"
	songWrappersPkg "github.com/K1tten2005/lyryx-backend/internal/usecases/song/wrappers"
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

var (
	userBucketName   = "user"
	artistBucketName = "artist"
	songBucketName   = "song"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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
	strictAuthMiddleware := echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(authHandlersPkg.JwtCustomClaims)
		},
		SigningKey: []byte(cfg.JWTSecret),
	})

	optionalAuthMiddleware := echojwt.WithConfig(echojwt.Config{
    NewClaimsFunc: func(c echo.Context) jwt.Claims {
        return new(authHandlersPkg.JwtCustomClaims)
    },
    SigningKey: []byte(cfg.JWTSecret),

    ContinueOnIgnoredError: true,
    ErrorHandler: func(c echo.Context, err error) error {
        return nil
    },
})
	checkRoleMiddleware := middlewares.NewRoleCheckerMiddleware(logger)

	echoHandler := echo.New()

	// Validator.
	echoHandler.Validator = utils.NewHTTPRequestValidator()

	// Claims getter.
	claimsGetter := auth.ClaimsGetter{}

	authStorage := authStoragePkg.NewStorage(db)
	authWrappers := authWrappersPkg.NewStorage(authStorage)
	authUsecase := authUsecasePkg.NewUsecase(authWrappers, &authUsecasePkg.BcryptHasher{})
	authHandlers := authHandlersPkg.New(authUsecase, []byte(cfg.JWTSecret), logger)
	authHandlers.RegisterHandlers(echoHandler, strictAuthMiddleware)

	userStorage := userStoragePkg.NewStorage(db, logger)
	userAvatarStorage := userStoragePkg.NewMinIOAvatarStorage(minioClient, userBucketName, cfg.MinIOPublicBaseURL)
	if err := userAvatarStorage.EnsureBucketPublic(context.Background()); err != nil {
		log.Errorf("failed ensure minio bucket public policy: %v", err)
		return
	}
	userWrappers := userWrappersPkg.NewStorage(userStorage)
	userAvatarWrapper := userWrappersPkg.NewUserAvatarStorage(userAvatarStorage)
	userUsecase := userUsecasePkg.NewUsecase(userWrappers, userAvatarWrapper, logger)
	userHandlers := userHandlersPkg.NewUserHandlers(userUsecase, claimsGetter, logger)
	userHandlers.RegisterHandlers(echoHandler, strictAuthMiddleware)

	artistStorage := artistStoragePkg.NewStorage(db, logger)
	artistAvatarStorage := artistStoragePkg.NewMinIOAvatarStorage(minioClient, artistBucketName, cfg.MinIOPublicBaseURL)
	if err := artistAvatarStorage.EnsureBucketPublic(context.Background()); err != nil {
		log.Errorf("failed ensure minio bucket public policy: %v", err)
		return
	}
	artistWrappers := artistWrappersPkg.NewStorage(artistStorage)
	artistAvatarWrapper := artistWrappersPkg.NewArtistAvatarStorage(artistAvatarStorage)
	artistUsecase := artistUsecasePkg.NewUsecase(artistWrappers, artistAvatarWrapper, logger)
	artistHandlers := artistHandlersPkg.NewArtistHandlers(artistUsecase, claimsGetter, logger)
	artistHandlers.RegisterHandlers(echoHandler, strictAuthMiddleware, checkRoleMiddleware)

	songStorage := songStoragePkg.NewStorage(db, logger)
	songCoverStorage := songStoragePkg.NewMinIOCoverStorage(minioClient, songBucketName, cfg.MinIOPublicBaseURL)
	if err := songCoverStorage.EnsureBucketPublic(context.Background()); err != nil {
		log.Errorf("failed ensure minio bucket public policy: %v", err)
		return
	}
	songWrappers := songWrappersPkg.NewStorage(songStorage)
	songCoverWrapper := songWrappersPkg.NewSongCoverStorage(songCoverStorage)
	songUsecase := songUsecasePkg.NewUsecase(songWrappers, songCoverWrapper, logger)
	songHandlers := songHandlersPkg.NewSongHandlers(songUsecase, claimsGetter, logger)
	songHandlers.RegisterHandlers(echoHandler, strictAuthMiddleware, checkRoleMiddleware)

	annotationStorage := annotationStoragePkg.NewStorage(db, logger)
	annotationWrappers := annotationWrappersPkg.NewStorage(annotationStorage)
	annotationUsecase := annotationUsecasePkg.NewUsecase(annotationWrappers, songUsecase, userUsecase, logger)
	annotationHandlers := annotationHandlersPkg.NewHandlers(annotationUsecase, claimsGetter, logger)
	annotationHandlers.RegisterHandlers(echoHandler, strictAuthMiddleware, optionalAuthMiddleware, checkRoleMiddleware)

	searchStorage := searchStoragePkg.NewStorage(db, logger)
	searchWrappers := searchWrappersPkg.NewStorage(searchStorage)
	searchUsecase := searchUsecasePkg.NewUsecase(searchWrappers, logger)
	searchHandlers := searchHandlersPkg.NewHandlers(searchUsecase, logger)
	searchHandlers.RegisterHandlers(echoHandler, strictAuthMiddleware, checkRoleMiddleware)

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
