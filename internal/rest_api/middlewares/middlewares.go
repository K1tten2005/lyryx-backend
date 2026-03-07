package middlewares

import (
	"github.com/K1tten2005/lyryx-backend/internal/model/roles"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

type RolesCheckerMiddleware struct {
	logger *log.Logger
}

func NewRoleCheckerMiddleware(logger *log.Logger) *RolesCheckerMiddleware {
	return &RolesCheckerMiddleware{logger: logger}
}

func (rcm *RolesCheckerMiddleware) CheckRole(minRole roles.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := auth.GetClaims(c)
			if err != nil {
				rcm.logger.WithError(err).Warning("[check_roles] get claims failed")
				return next(c)
			}

			// Проверяем: уровень пользователя >= минимально нужного?
			if roles.RoleLevel[roles.Role(claims.Role)] < roles.RoleLevel[minRole] {
				rcm.logger.WithField("role", claims.Role).Warning("role check failed")
				return next(c)
			}

			return next(c)
		}
	}
}
