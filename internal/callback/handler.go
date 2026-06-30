package callback

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/dispatcher"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/response"
)

// Handler handles callback requests.
// dispatcher: The dispatcher for handling management and flow actions.
func Handler(dispatcher *dispatcher.Dispatcher) echo.HandlerFunc {
	return func(c echo.Context) error {
		logrus.Debugf("received callback request: %s", c.Request().URL.Path)

		ctx := c.Request().Context()

		requestID := c.Param("requestId")

		defer c.Request().Body.Close()
		data, bErr := io.ReadAll(c.Request().Body)
		if bErr != nil {
			logrus.Errorf(
				"failed to read callback request body for request %s: %v",
				requestID,
				bErr,
			)
			return echo.NewHTTPError(http.StatusBadRequest, response.NewError(response.ErrInvalidPayload))
		}
		body := json.RawMessage(data)
		logrus.Debugf("callback request body: %s", string(body))

		return dispatcher.HandleCallback(ctx, c, requestID, body)
	}
}
