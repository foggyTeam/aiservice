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
			// Skip auth in development environment
			if cfg.Server.Env == "dev" {
				return next(c)
			}

			// Get the API key from the header
			apiKey := c.Request().Header.Get("x-api-key")
			
			// Check if the API key matches the expected verification key
			expectedKey := cfg.Server.VerificationKey
			
			if apiKey != expectedKey {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized: Invalid API key",
				})
			}

			// API key is valid, continue with the request
			return next(c)
		}
	}
}