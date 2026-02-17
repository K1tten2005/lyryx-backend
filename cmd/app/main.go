package main

import (
	"github.com/K1tten2005/lyryx-backend/internal/config"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api"
	authHandlers "github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/utils"
	authUsecase "github.com/K1tten2005/lyryx-backend/internal/usecases/auth"
	authStorage "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/storage"
	authWrappers "github.com/K1tten2005/lyryx-backend/internal/usecases/auth/wrappers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

// @title lyryx API
// @version 1.0
// @host localhost:8080
func main() {
	// Logger.
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})

	// Config.
	var cfg config.Config
	err := env.Parse(&cfg)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %+v", r)
		}
	}()

	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Errorf("failed sqlx connect: %v", err)
		return
	}

	authConfig := echojwt.Config{
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(authHandlers.JwtCustomClaims)
		},
		SigningKey: cfg.JWTSecret,
	}

	echoHandler := echo.New()

	// Auth middleware.
	authMiddleware := echojwt.WithConfig(authConfig)

	// Validator.
	echoHandler.Validator = utils.NewHTTPRequestValidator()

	authStorage := authStorage.NewStorage(db)
	authWrappers := authWrappers.NewStorage(authStorage)
	authUsecase := authUsecase.NewUsecase(authWrappers, &authUsecase.BcryptHasher{})
	authHandlers := authHandlers.New(authUsecase, []byte(cfg.JWTSecret), logger)
	authHandlers.RegisterHandlers(echoHandler, authMiddleware)

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
