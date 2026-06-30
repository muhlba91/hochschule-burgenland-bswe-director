//nolint:cyclop // This file is complex due to the nature of websocket handling.
package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// keepaliveInterval defines how often to send pings to the client to keep the connection alive.
const keepaliveInterval = 30 * time.Second

// pingTimeout defines how long to wait for a pong response before considering the connection dead.
const pingTimeout = 5 * time.Second

// Handler handles websocket requests.
// dispatcher: The dispatcher for handling management and flow actions.
//
//nolint:gocognit,funlen // handler is complex due to the nature of websocket handling.
func Handler(dispatcher *orchestrator.Dispatcher) echo.HandlerFunc {
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

		connData := connection.NewData()
		ctx := c.Request().Context()

		defer func() {
			handleDisconnect(ctx, dispatcher, conn, connData)
		}()

		readDone := make(chan struct{})
		msgCh := make(chan message.Message)

		go func() {
			defer close(readDone)
			for {
				var msg message.Message
				if errRead := wsjson.Read(ctx, conn, &msg); errRead != nil {
					if websocket.CloseStatus(errRead) == -1 {
						logrus.Errorf("websocket read failed: %v", errRead)
					} else {
						logrus.Infof("websocket closed by client: %v", errRead)
					}
					return
				}
				msgCh <- msg
			}
		}()

		var broadcastSub store.Subscription
		var broadcastCh <-chan string

		defer func() {
			if broadcastSub != nil {
				_ = broadcastSub.Close()
			}
		}()

		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()

		for {
			select {
			case msg := <-msgCh:
				logrus.Debugf("received message: %s", msg)
				sid := dispatcher.Handle(ctx, conn, connData, msg)

				if sid != nil && connData.SessionID != nil && *sid != *connData.SessionID {
					logrus.Debugf(
						"connection requested a new session, but already subscribed to a session with connection: %s, %s %s",
						*sid,
						*connData.SessionID,
						connData.ConnectionID,
					)
					handleDisconnect(ctx, dispatcher, conn, connData)
					_ = broadcastSub.Close()
					connData.SessionID = nil
					broadcastSub = nil
					broadcastCh = nil
				}
				if sid != nil && (connData.SessionID == nil || *sid != *connData.SessionID) {
					logrus.Debugf(
						"subscribing to new session with connection: %s, %s",
						*sid,
						connData.ConnectionID,
					)
					connData.SessionID = sid
					sub, subErr := dispatcher.Subscribe(ctx, *connData.SessionID)
					if subErr != nil {
						logrus.Errorf("failed to subscribe to session %s: %v", *connData.SessionID, subErr)
						continue
					}
					broadcastSub = sub
					broadcastCh = broadcastSub.Channel()
				}

			case msgPayload := <-broadcastCh:
				logrus.Debugf("received broadcast message: %s", msgPayload)
				if errWrite := wsjson.Write(ctx, conn, json.RawMessage(msgPayload)); errWrite != nil {
					logrus.Errorf("websocket write failed: %v", errWrite)
					continue
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

// handleDisconnect handles the disconnection of a websocket connection.
// ctx: The context for managing request lifecycle.
// dispatcher: The dispatcher for handling management and flow actions.
// conn: The websocket connection that is being disconnected.
// connectionData: The connection data associated with the websocket connection.
func handleDisconnect(
	ctx context.Context,
	dispatcher *orchestrator.Dispatcher,
	conn *websocket.Conn,
	connectionData *connection.Data,
) {
	if connectionData.SessionID != nil {
		logrus.Infof(
			"handling disconnect for session and connection: %s %s",
			*connectionData.SessionID,
			connectionData.ConnectionID,
		)

		payload, _ := json.Marshal(connectionData)
		dispatcher.Handle(ctx, conn, connectionData, message.Message{
			Event:   event.GenerateEventName(connectionData.GetFlowName(), event.Disconnected),
			Payload: payload,
		})
	}
}
