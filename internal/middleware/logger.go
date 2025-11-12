package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogHost:   true,
		LogError:  true,
		BeforeNextFunc: func(c echo.Context) {
			c.Set("logger", logger)
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("request",
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("host", v.Host),
					zap.Duration("latency", v.Latency),
				)
			} else {
				logger.Error("request error",
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("host", v.Host),
					zap.Duration("latency", v.Latency),
					zap.Error(v.Error),
				)
			}
			return nil
		},
	})
}
