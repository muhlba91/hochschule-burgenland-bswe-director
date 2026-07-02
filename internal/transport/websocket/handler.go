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

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
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
			logrus.WithFields(logrus.Fields{
				logging.FieldError: err,
			}).Error("failed to accept websocket connection")
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
						logrus.WithFields(logrus.Fields{
							logging.FieldError:     errRead,
							logging.FieldSessionID: connData.SessionID,
						}).Error("websocket read failed")
					} else {
						logrus.WithFields(logrus.Fields{
							logging.FieldError:     errRead,
							logging.FieldSessionID: connData.SessionID,
						}).Info("websocket closed by client")
					}
					return
				}
				msgCh <- msg
			}
		}()

		var broadcastSub store.Subscription
		var broadcastCh <-chan string
		sessionUpdateCh := make(chan *string)

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
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: connData.SessionID,
					logging.FieldEvent:     msg.Event,
				}).Debug("received websocket message")
				go func(m message.Message) {
					sid := dispatcher.Handle(ctx, conn, connData, m)
					if sid != nil {
						sessionUpdateCh <- sid
					}
				}(msg)

			case sid := <-sessionUpdateCh:
				if connData.SessionID != nil && *sid != *connData.SessionID {
					logrus.WithFields(logrus.Fields{
						"new_session_id":          *sid,
						"old_session_id":          *connData.SessionID,
						logging.FieldConnectionID: connData.ConnectionID,
					}).Debug("connection requested a new session, but already subscribed to a session")
					handleDisconnect(ctx, dispatcher, conn, connData)
					if broadcastSub != nil {
						_ = broadcastSub.Close()
					}
					connData.SessionID = nil
					broadcastSub = nil
					broadcastCh = nil
				}
				if connData.SessionID == nil || *sid != *connData.SessionID {
					logrus.WithFields(logrus.Fields{
						logging.FieldSessionID:    *sid,
						logging.FieldConnectionID: connData.ConnectionID,
					}).Debug("subscribing to new session")
					connData.SessionID = sid
					sub, subErr := dispatcher.Subscribe(ctx, *connData.SessionID)
					if subErr != nil {
						logrus.WithFields(logrus.Fields{
							logging.FieldSessionID: *connData.SessionID,
							logging.FieldError:     subErr,
						}).Error("failed to subscribe to session")
						continue
					}
					broadcastSub = sub
					broadcastCh = broadcastSub.Channel()
				}

			case msgPayload := <-broadcastCh:
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: connData.SessionID,
				}).Debug("received broadcast message")
				if errWrite := wsjson.Write(ctx, conn, json.RawMessage(msgPayload)); errWrite != nil {
					logrus.WithFields(logrus.Fields{
						logging.FieldSessionID: connData.SessionID,
						logging.FieldError:     errWrite,
					}).Error("websocket write failed")
					continue
				}

			case <-readDone:
				return nil

			case <-ticker.C:
				pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
				errPing := conn.Ping(pingCtx)
				cancel()
				if errPing != nil {
					logrus.WithFields(logrus.Fields{
						logging.FieldSessionID: connData.SessionID,
						logging.FieldError:     errPing,
					}).Error("websocket ping failed")
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
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID:    *connectionData.SessionID,
			logging.FieldConnectionID: connectionData.ConnectionID,
		}).Info("handling disconnect for session and connection")

		payload, _ := json.Marshal(connectionData)
		dispatcher.Handle(ctx, conn, connectionData, message.Message{
			Event:   event.GenerateEventName(connectionData.GetFlowName(), event.Disconnected),
			Payload: payload,
		})
	}
}
