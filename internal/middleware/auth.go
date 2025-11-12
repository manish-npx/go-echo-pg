package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/manish-npx/go-echo-pg/internal/config"
	"github.com/manish-npx/go-echo-pg/internal/utils"
)

func AuthMiddleware(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]
			claims, err := utils.ValidateToken(tokenString, config)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set("userID", claims.UserID)
			c.Set("userEmail", claims.Email)

			return next(c)
		}
	}
}
