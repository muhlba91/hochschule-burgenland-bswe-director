# Pokémon Flow Events & Actions

This document outlines the WebSocket events and HTTP callback actions implemented for the turn-based Pokémon card game flow.

For detailed structure types and fields, please consult the source code under the `pkg/flows/pokemon/` package.

---

## WebSocket Events (Client <-> Director)

The frontend client communicates with the Pokémon flow using real-time JSON events sent over the WebSocket connection.

Path: `pkg/flows/pokemon/event/`

- **`pokemon:create`**: Sent by a client to initialize a new game room. Payload defines names and callback URLs for both players. Responds with **`pokemon:created`** returning the generated `sessionId`.

- **`pokemon:join`**: Dispatched by a player to bind their WebSocket connection to an active game lobby. Responds with **`pokemon:joined`** upon successful lobby registration.

- **`pokemon:list`**: Queries all active game sessions in the Redis store. Responds with **`pokemon:listing`** containing a dictionary of current game sessions.

- **`pokemon:state`**: Requests a comprehensive snapshot of the game board. Responds with **`pokemon:current_state`** containing the active board details.

- **`pokemon:player_information`**: Queries player connection metadata and cards. Responds with **`pokemon:player_details`** including player-specific filtered card views and connection information.

---

## Callback Actions (Director <-> Agent)

The Flow Director coordinates gameplay actions asynchronously by posting payload structures to each player's service URL, expecting replies on a unique `/api/v1/callback/:requestId` endpoint.

Path: `pkg/flows/pokemon/action/`

- **`pokemon:start`**: Sent concurrently to both players to initialize their deck configurations. The agent responds by returning their chosen starting deck state.

- **`pokemon:turn`**: Dispatched to the player whose turn is currently active. The request payload contains the player-specific filtered view of the board. The agent decides their moves and replies with a turn result (including actions, card placements, or passing details).

- **`pokemon:attack`**: Sent to coordinate combat execution when an attack is selected. The request payload contains the chosen attack and target details. The opponent's service is notified of damage and card status adjustments.
