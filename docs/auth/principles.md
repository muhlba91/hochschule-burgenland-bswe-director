# Authentication Principles

This document describes the authentication mechanisms used in the Flow Director to ensure secure communication between the Director, Agents (Clients), and the Frontend.

## Overview

There are three main communication paths that require authentication:

1. **Frontend/Client to Director (WebSocket):** Uses short-lived JWTs.
2. **Handshake (Token Exchange):** Uses a **Shared Secret** (Team Secret) to generate an HMAC-based API Key.
3. **Callback (Request/Response):** Uses a unique **Request Secret** (sent by the Director in `X-Callback-Secret`) to sign response payloads via an `X-Signature` header.

---

## 1. API Key & JWT (WebSocket Authentication)

The WebSocket connection requires a JWT for authentication. Since the frontends are typically public-facing, they should not store a long-lived secret. Instead, they perform a **Token Exchange**.

### The Handshake

1. **Client ID & Team Secret:** Every student team/flow agent has a unique `clientID` and a long-lived **Team Secret** provided by the instructor.
2. **API Key Generation (HMAC):**
   The `apiKey` is an HMAC-SHA256 hash of the `clientID` using the **Team Secret** as the key.
   $$apiKey = HMAC-SHA256(TeamSecret, clientID)$$
3. **Token Exchange:**
   The client sends a POST request to `/api/v1/auth/token` with:

   ```json
   {
     "clientID": "team-01",
     "apiKey": "..." 
   }
    ```

4. **JWT Retrieval:**
   The Director verifies the API key and returns a short-lived JWT.
5. **WebSocket Connection:**
   The client connects to the WebSocket endpoint (e.g., via query parameter or header, depending on implementation).

---

## 2. Callback Authentication (HMAC & Secrets)

When the Director calls a Flow Agent (e.g., to ask for a move), it needs to ensure only the authorized agent responds, and the agent needs to ensure the request came from the Director.

### Director -> Engine (Outgoing Request)

* The Director sends an HTTP POST to the Engine's endpoint.
* If `CallbackAuthEnabled` is true, the Director generates a unique **Secret** for **that specific request**.
* This secret is sent in the HTTP Header: `X-Callback-Secret`.

### Engine -> Director (Incoming Callback Response)

To respond to a request, the Engine must sign the response body using the provided secret:

1. **POST** the response to the `callbackURL` provided in the request body.
2. **Generate a Signature**: Create an HMAC-SHA256 hash of the **raw response body** using the `X-Callback-Secret` as the key.
   $$X-Signature = HMAC-SHA256(X-Callback-Secret, response\_body)$$
3. **Authentication Header**: Include this signature in the `X-Signature` HTTP header (hex-encoded).

---

## 3. Security Considerations

* **Secret Management:** The `secret` used for HMAC must never be exposed in frontend code. If an Angular app needs to connect, it should ideally fetch the JWT via a backend proxy or use a secure environment-specific configuration if it's a known internal client.
* **TTL:** JWTs are short-lived. Clients must be prepared to refresh them if the WebSocket connection drops or the token expires.
* **HMAC Comparison:** Always use a constant-time comparison (like `hmac.Equal` in Go) to prevent timing attacks when verifying keys.
