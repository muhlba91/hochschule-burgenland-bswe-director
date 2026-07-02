//nolint:cyclop // This file is complex due to the nature of websocket handling.
package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo/v4"

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
			slog.ErrorContext(c.Request().Context(), "failed to accept websocket connection",
				slog.Any(logging.FieldError, err),
			)
			return err
		}
		defer conn.Close(websocket.StatusInternalError, "")

		connData := connection.NewData()
		ctx := c.Request().Context()

		defer func() {
			handleDisconnect(context.WithoutCancel(ctx), dispatcher, connData)
		}()

		readDone := make(chan struct{})
		msgCh := make(chan message.Message)

		go func() {
			defer close(readDone)
			for {
				var msg message.Message
				if errRead := wsjson.Read(ctx, conn, &msg); errRead != nil {
					if websocket.CloseStatus(errRead) == -1 {
						slog.ErrorContext(ctx, "websocket read failed",
							slog.Any(logging.FieldError, errRead),
							slog.Any(logging.FieldSessionID, connData.SessionID),
						)
					} else {
						slog.InfoContext(ctx, "websocket closed by client",
							slog.Any(logging.FieldError, errRead),
							slog.Any(logging.FieldSessionID, connData.SessionID),
						)
					}
					return
				}
				msgCh <- msg
			}
		}()

		var broadcastSub store.Subscription
		var broadcastCh <-chan string
		sessionUpdateCh := make(chan *string)
		writeCh := make(chan *message.Message)

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
				slog.DebugContext(ctx, "received websocket message",
					slog.Any(logging.FieldSessionID, connData.SessionID),
					slog.String(logging.FieldEvent, msg.Event),
				)
				go func(m message.Message) {
					sid, rMsg := dispatcher.Handle(ctx, connData, m)
					if rMsg != nil {
						writeCh2 := writeCh
						writeCh2 <- rMsg
					}
					if sid != nil {
						sessionUpdateCh <- sid
					}
				}(msg)

			case sid := <-sessionUpdateCh:
				if connData.SessionID != nil && *sid != *connData.SessionID {
					slog.DebugContext(ctx, "connection requested a new session, but already subscribed to a session",
						slog.String("new_session_id", *sid),
						slog.String("old_session_id", *connData.SessionID),
						slog.String(logging.FieldConnectionID, connData.ConnectionID),
					)
					handleDisconnect(ctx, dispatcher, connData)
					if broadcastSub != nil {
						_ = broadcastSub.Close()
					}
					connData.SessionID = nil
					broadcastSub = nil
					broadcastCh = nil
				}
				if connData.SessionID == nil || *sid != *connData.SessionID {
					slog.DebugContext(ctx, "subscribing to new session",
						slog.String(logging.FieldSessionID, *sid),
						slog.String(logging.FieldConnectionID, connData.ConnectionID),
					)
					connData.SessionID = sid
					sub, subErr := dispatcher.Subscribe(ctx, *connData.SessionID)
					if subErr != nil {
						slog.ErrorContext(ctx, "failed to subscribe to session",
							slog.String(logging.FieldSessionID, *connData.SessionID),
							slog.Any(logging.FieldError, subErr),
						)
						continue
					}
					broadcastSub = sub
					broadcastCh = broadcastSub.Channel()
				}

			case rMsg := <-writeCh:
				slog.DebugContext(ctx, "received response message to write",
					slog.Any(logging.FieldSessionID, connData.SessionID),
				)
				if errWrite := wsjson.Write(ctx, conn, rMsg); errWrite != nil {
					slog.ErrorContext(ctx, "websocket write failed for response message",
						slog.Any(logging.FieldSessionID, connData.SessionID),
						slog.Any(logging.FieldError, errWrite),
					)
					continue
				}

			case msgPayload := <-broadcastCh:
				slog.DebugContext(ctx, "received broadcast message",
					slog.Any(logging.FieldSessionID, connData.SessionID),
				)
				if errWrite := wsjson.Write(ctx, conn, json.RawMessage(msgPayload)); errWrite != nil {
					slog.ErrorContext(ctx, "websocket write failed",
						slog.Any(logging.FieldSessionID, connData.SessionID),
						slog.Any(logging.FieldError, errWrite),
					)
					continue
				}

			case <-readDone:
				return nil

			case <-ticker.C:
				pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
				errPing := conn.Ping(pingCtx)
				cancel()
				if errPing != nil {
					slog.ErrorContext(ctx, "websocket ping failed",
						slog.Any(logging.FieldSessionID, connData.SessionID),
						slog.Any(logging.FieldError, errPing),
					)
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
// connectionData: The connection data associated with the websocket connection.
func handleDisconnect(
	ctx context.Context,
	dispatcher *orchestrator.Dispatcher,
	connectionData *connection.Data,
) {
	if connectionData.SessionID != nil {
		slog.InfoContext(ctx, "handling disconnect for session and connection",
			slog.String(logging.FieldSessionID, *connectionData.SessionID),
			slog.String(logging.FieldConnectionID, connectionData.ConnectionID),
		)

		payload, _ := json.Marshal(connectionData)
		_, _ = dispatcher.Handle(ctx, connectionData, message.Message{
			Event:   event.GenerateEventName(connectionData.GetFlowName(), event.Disconnected),
			Payload: payload,
		})
	}
}
