package middleware

import (
	"net/http"

	"github.com/aiservice/internal/config"
	"github.com/labstack/echo/v4"
)

// APIKeyAuth creates middleware that validates API key in request headers
func APIKeyAuth(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Server.Env == "dev" {
				return next(c)
			}

			apiKey := c.Request().Header.Get("x-api-key")
			
			expectedKey := cfg.Server.VerificationKey
			
			if apiKey != expectedKey {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized: Invalid API key",
				})
			}

			return next(c)
		}
	}
}