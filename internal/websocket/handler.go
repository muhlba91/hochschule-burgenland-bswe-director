//nolint:cyclop // Package websocket provides the websocket handler for the director application.
package websocket

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/websocket/dispatcher"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// keepaliveInterval defines how often to send pings to the client to keep the connection alive.
const keepaliveInterval = 30 * time.Second

// pingTimeout defines how long to wait for a pong response before considering the connection dead.
const pingTimeout = 5 * time.Second

// Handler handles websocket requests.
// dispatcher: The dispatcher for handling management and flow actions.
//
//nolint:gocognit // handler is complex due to the nature of websocket handling.
func Handler(dispatcher *dispatcher.Dispatcher) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			OriginPatterns:     []string{"*"},
		})
		if err != nil {
			logrus.Errorf("failed to accept websocket connection: %v", err)
			return err
		}
		defer conn.Close(websocket.StatusInternalError, "")

		ctx := c.Request().Context()

		readDone := make(chan struct{})
		msgCh := make(chan message.Message)

		go func() {
			defer close(readDone)
			for {
				var msg message.Message
				if errRead := wsjson.Read(ctx, conn, &msg); errRead != nil {
					logrus.Errorf("websocket read failed: %v", errRead)
					return
				}
				msgCh <- msg
			}
		}()

		var sub *redis.PubSub
		var redisCh <-chan *redis.Message

		var sessionID string
		defer func() {
			if sub != nil {
				_ = sub.Close()
			}
		}()

		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()

		for {
			select {
			case msg := <-msgCh:
				sid := dispatcher.Handle(ctx, conn, sessionID, msg)

				if sid != nil {
					sessionID = *sid
					sub = dispatcher.Subscribe(ctx, sessionID)
					redisCh = sub.Channel()
				}

			case redisMsg := <-redisCh:
				if errWrite := wsjson.Write(ctx, conn, redisMsg.Payload); errWrite != nil {
					logrus.Errorf("websocket write failed: %v", errWrite)
					return nil
				}

			case <-readDone:
				return nil

			case <-ticker.C:
				pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
				errPing := conn.Ping(pingCtx)
				cancel()
				if errPing != nil {
					logrus.Errorf("websocket ping failed: %v", errPing)
					return nil
				}

			case <-ctx.Done():
				_ = conn.Close(websocket.StatusNormalClosure, "")
				return nil
			}
		}
	}
}
