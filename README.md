# Hochschule Burgenland - BSWE - Director

[![](https://img.shields.io/github/license/muhlba91/hochschule-burgenland-bswe-director?style=for-the-badge)](LICENSE.md)
[![](https://img.shields.io/github/actions/workflow/status/muhlba91/hochschule-burgenland-bswe-director/verify.yml?style=for-the-badge)](https://github.com/muhlba91/hochschule-burgenland-bswe-director/actions/workflows/verify.yml)
[![](https://img.shields.io/coverallsCoverage/github/muhlba91/hochschule-burgenland-bswe-director?style=for-the-badge)](https://github.com/muhlba91/hochschule-burgenland-bswe-director/)
[![](https://api.scorecard.dev/projects/github.com/muhlba91/hochschule-burgenland-bswe-director/badge?style=for-the-badge)](https://scorecard.dev/viewer/?uri=github.com/muhlba91/hochschule-burgenland-bswe-director)
[![](https://img.shields.io/github/release-date/muhlba91/hochschule-burgenland-bswe-director?style=for-the-badge)](https://github.com/muhlba91/hochschule-burgenland-bswe-director/releases)
[![](https://img.shields.io/github/all-contributors/muhlba91/hochschule-burgenland-bswe-director?color=ee8449&style=for-the-badge)](#contributors)
<a href="https://www.buymeacoffee.com/muhlba91" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="28" width="150"></a>

**Flow Director** is a generic, high-concurrency Go-based service designed to orchestrate state-machine flows. It serves as a central middleware hub, bridging real-time clients (WebSockets) and external player/agent microservices (HTTP POST callbacks) using a highly modular registry.

---

## Architecture Overview

### Key Components

- **Registry**: Allows developer-defined flows to register custom WebSocket events and Callback action handlers.
- **Requestor**: Dispatches async requests to external player URLs, manages retry limits, handles request parallelization constraints, and tracks pending callbacks in Redis.
- **Distributed Locking**: Employs Redis-based distributed locks to protect state transitions from concurrent WebSockets or HTTP callback events.
- **Real-time Pub/Sub**: Broadcasts state updates to all active client WebSocket connections automatically when session data changes.

### Communication Paradigms

The director processes two major categories of communication to coordinate active sessions:

#### WebSocket Messages (Client <-> Director)

Clients connect via the WebSocket endpoint.

- **Inbound (Client to Director)**: Used for session administration, including creating lobbies, joining sessions, querying session directories, or manual state retrieval.
- **Outbound (Director to Client)**: Broadcasts updated, real-time board states, connection events, or winner declarations to all active connections in a session whenever data transitions.

#### Callback Actions & Requests (Director <-> Agent)

Communication with player or automated agent microservices occurs asynchronously via HTTP.

- **Outgoing Request (Director to Agent)**: When the state machine requires an action (e.g., flow actions), the **Requestor** dispatches a POST request containing the current filtered session state and a unique callback URL: `<BASE_URL>/callback/<requestId>`.
- **Inbound Callback (Agent to Director)**: Once the agent determines its choice, it posts the result back to the unique callback route, prompting the dispatcher to lock, apply, and broadcast the transaction.

---

## Supported Flows

The director is designed to be easily extensible. 

### Pokémon Game Flow

The director supports a turn-based Pokémon card game flow. It registers game logic to handle:

- **Session Initialization**: Automatically provisions player decks, HP trackers, and energy cards.
- **Sequential Turn Coordination**: Coordinates player hand state filtering and sequential decision-making requests.
- **Action & Damage Parsing**: Executes selected attacks, resolves damage modifiers, checks state for defeated cards, and evaluates winning conditions.

---

## Configuration & Deployment

### Environment Configuration

| Variable                  | Description                                                        | Default                 |
| ------------------------- | ------------------------------------------------------------------ | ----------------------- |
| `SERVER_HOST`             | Host address for the Echo API server to listen on.                 | `0.0.0.0`               |
| `SERVER_PORT`             | Port for the Echo API server.                                      | `8888`                  |
| `HEALTHZ_HOST`            | Host address for Kubernetes health probe endpoints.                | `0.0.0.0`               |
| `HEALTHZ_PORT`            | Port for health probe endpoints.                                   | `8080`                  |
| `REDIS_HOST`              | Redis database address.                                            | `localhost`             |
| `REDIS_PORT`              | Redis database port.                                               | `6379`                  |
| `REDIS_PASSWORD`          | Redis authentication password (optional).                          | *(empty)*               |
| `BASE_URL`                | Public URL of this Director, used for generating callback links.   | `http://localhost:8888` |
| `RATE_LIMIT`              | Number of allowed requests per second.                             | `10`                    |
| `BURST_LIMIT`             | Maximum number of requests allowed in a burst.                     | `20`                    |
| `RATE_LIMIT_WINDOW`       | Duration for which the rate limit is applied.                      | `3m`                    |
| `CALLBACK_AUTH_ENABLED`   | Enable HMAC verification for incoming callbacks.                   | `false`                 |
| `AUTH_ENABLED`            | Enable JWT authentication for WebSocket connections.               | `false`                 |
| `AUTH_SECRET`             | Secret for signing JWT tokens and verifying API keys.              | *(empty)*               |
| `TOKEN_TTL`               | JWT token expiration time in seconds.                              | `86400`                  |

### Development

Build and test the application using the included commands:

```shell
make lint        # Run linter checks
make fix         # Format source code
make build       # Compile binary
make test        # Run unit tests
make coverage    # Generate code coverage reports
```

### Deployment

#### Docker

Run a local Redis (or Valkey) container:

> [!IMPORTANT]
> The Director requires **Redis 7.0+** or **Valkey 7.0+** for distributed locking and state management features.

```shell
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

Run the Director:

```shell
docker run -d \
  --name director \
  -p 8888:8888 \
  -p 8080:8080 \
  -e REDIS_HOST="host.docker.internal" \
  -e BASE_URL="http://localhost:8888" \
  ghcr.io/muhlba91/hochschule-burgenland-bswe-director:latest
```

#### Helm

The Helm chart is located in the [`charts/director/`](charts/director/) folder and is hosted in the GitHub Container Registry (GHCR).

```bash
helm upgrade --install director oci://ghcr.io/muhlba91/hochschule-burgenland-bswe-director-charts/director
```

The chart supports both standard Kubernetes Ingress and the Gateway API (`HTTPRoute` & `ListenerSet`).

> [!IMPORTANT]
> This chart requires an external Redis or Valkey instance. You can provide connection details via `values.yaml` or an existing secret.

##### Ingress

Enable ingress in `values.yaml`:

```yaml
ingress:
  enabled: true
  className: "traefik"
  hosts:
    - host: director.example.com
      paths:
        - path: /
          pathType: ImplementationSpecific
```

##### Gateway API

Enable Gateway API support in `values.yaml`:

```yaml
gateway:
  enabled: true
  gatewayClassName: "my-gateway"
  hosts:
    - host: director.example.com
      path: /
  listenerSet:
    enabled: true
    listeners:
      - name: http
        port: 80
        protocol: HTTP
        allowedRoutes:
          namespaces:
            from: Same
```

---

## Documentation

Comprehensive documentation for the project is located in the [`docs/`](docs/) directory:

- **[API Specification](docs/openapi.yaml)**: OpenAPI/Swagger definition of the Director's HTTP endpoints.
- **[WebSocket Events](docs/generic-websocket-events.md)**: Details on the generic events transmitted over the WebSocket connection.
- **[Authentication](docs/auth/principles.md)**: Principles and implementation details for securing communications, including [Angular](docs/auth/samples/angular-auth.service.ts) and [Java](docs/auth/samples/FlowCallbackController.java) samples.
- **[Flows](docs/flows/)**: Detailed documentation for supported flows, such as the [Pokémon Game Flow](docs/flows/pokemon.md).

---

## Health Probes

The director exposes dedicated health endpoints for container orchestrators:

- **Liveness**: `/livez` - Verifies the server is operational.
- **Readiness**: `/healthz` - Tests connections to backend Redis stores. Returns `200 OK` or `503 Service Unavailable`.
- **Startup**: `/startupz` - Asserts successful server startup.
