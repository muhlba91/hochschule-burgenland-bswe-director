package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
)

// Handler returns an echo handler for the token exchange endpoint.
// cfg: The configuration data containing the auth secret and token TTL.
func Handler(cfg *configuration.Data) echo.HandlerFunc {
	return func(c echo.Context) error {
		type tokenReq struct {
			ClientID string `json:"clientID"`
			APIKey   string `json:"apiKey"`
		}
		var req tokenReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, response.NewError(ErrInvalidRequest))
		}

		if !VerifyAPIKey(req.ClientID, req.APIKey, cfg.AuthSecret) {
			return c.JSON(http.StatusUnauthorized, response.NewError(ErrInvalidCredentials))
		}

		token, err := GenerateJWT(req.ClientID, cfg.AuthSecret, cfg.TokenTTL)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.NewError(ErrSecretNotConfigured))
		}

		return c.JSON(http.StatusOK, map[string]any{
			"token":      token,
			"expires_in": cfg.TokenTTL,
		})
	}
}
