package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aiservice/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name             string
		env              string
		verificationKey  string
		requestKey       string
		expectedStatus   int
	}{
		{
			name:             "Dev environment - should bypass auth",
			env:              "dev",
			verificationKey:  "any-key",
			requestKey:       "",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Prod environment - valid key",
			env:              "prod",
			verificationKey:  "valid-key",
			requestKey:       "valid-key",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Prod environment - invalid key",
			env:              "prod",
			verificationKey:  "valid-key",
			requestKey:       "invalid-key",
			expectedStatus:   http.StatusUnauthorized,
		},
		{
			name:             "Prod environment - no key provided",
			env:              "prod",
			verificationKey:  "valid-key",
			requestKey:       "",
			expectedStatus:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			
			// Create config with test values
			cfg := &config.Config{
				Server: config.ServerConfig{
					Env:            tt.env,
					VerificationKey: tt.verificationKey,
				},
			}
			
			// Create a simple handler that returns OK
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}
			
			// Apply the middleware
			middlewareHandler := APIKeyAuth(cfg)(handler)
			
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.requestKey != "" {
				req.Header.Set("x-api-key", tt.requestKey)
			}
			
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			
			// Call the handler
			err := middlewareHandler(c)
			assert.NoError(t, err)
			
			// Check the response status
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			// If unauthorized, check the error message
			if tt.expectedStatus == http.StatusUnauthorized {
				var response map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Unauthorized")
			}
		})
	}
}