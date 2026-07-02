# Generic WebSocket Events

This document provides an overview of the generic WebSocket messaging structure and connection events used by the Flow Director.

For exact implementation details, please refer to the source code under the `pkg/transport/websocket/` package.

---

## Message Envelope

All real-time communications over the WebSocket connection utilize a standard JSON envelope. This ensures that the dispatcher can parse the event type before delegating the payload to a registered flow handler.

Path: `pkg/transport/websocket/message/message.go`

```json
{
  "event": "string",
  "payload": {}
}
```

- **`event`**: A string indicating the routing namespace and action type (formatted as `<flow_name>:<event_type>`).

- **`payload`**: A raw JSON object containing the specific event parameters.

---

## Connection Data

When a client establishes or updates a WebSocket session, connection-specific metadata is generated and tracked by the orchestrator.

Path: `pkg/transport/websocket/connection/data.go`

- **`connection_id`**: A unique UUID string generated upon accepting the WebSocket connection.

- **`session_id`**: An optional string identifying the game or flow session associated with the connection.

- **`internal_id`**: An optional internal string identifying the specific player representation within the session.

---

## Global Event Types

Global events represent system-level state or connection state transitions. They are generated using the utility function in `pkg/transport/websocket/event/global.go` and take the format `<flowName>:<eventType>`.

Path: `pkg/transport/websocket/event/global.go`

- **`connection:information`**: Emitted to share system connection metrics and connection IDs with the client.

- **`connection:disconnect`**: Dispatched when a client requests a manual disconnection or triggers a timeout.

- **`connection:disconnected`**: Broadcasted to session participants when a client's socket is severed or closed.

---

## Error Events

When an operation or message parsing fails, the director sends an error event back to the client.

Path: `pkg/transport/websocket/message/error.go`

- **Event Name**: `"error"`

- **Payload**: A simple JSON string representing the error reason (e.g., `"session full"`, `"session not found"`, `"unexpected message type"`).
