package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/manish-npx/go-echo-pg/internal/config"
)

func CORS(config *config.Config) echo.MiddlewareFunc {
	allowedOrigins := strings.Split(config.CORS.AllowedOrigins, ",")

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
