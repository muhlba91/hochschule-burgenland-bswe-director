package orchestrator

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// httpTimeout defines the timeout duration for HTTP requests made by the Requestor.
const httpTimeout = 5 * time.Second

// maxRetries defines the maximum number of retry attempts for sending a request in case of failure.
const maxRetries = 3

// requestWaitTime defines the wait time between retry attempts when sending a request.
const maxRequestWaitTime = 1 * time.Minute

// callbackTimeout defines the maximum duration to wait for a callback response before timing out.
const callbackTimeout = 10 * time.Minute

// pollInterval defines the interval at which the requestor checks for a callback response.
const pollInterval = 10 * time.Second

// RequestBuilderFunc is a function type that takes a session ID and a session object, and returns a callback request.
type RequestBuilderFunc func(string) (*callback.Request, callback.RequestData)

// Requestor is responsible for handling requests related to the Pokémon game flow.
type Requestor struct {
	requestorStore store.RequestStore
	sessionStore   store.SessionStore
	broadcastStore store.BroadcastStore
	httpClient     *http.Client
	config         *configuration.Data
}

// NewRequestor creates a new instance of Requestor.
// requestorStore: The store for managing requests.
// sessionStore: The store for managing sessions.
// broadcastStore: The store for managing broadcasts.
// config: The configuration data for the requestor.
func NewRequestor(
	requestorStore store.RequestStore,
	sessionStore store.SessionStore,
	broadcastStore store.BroadcastStore,
	config *configuration.Data,
) *Requestor {
	return &Requestor{
		requestorStore: requestorStore,
		sessionStore:   sessionStore,
		broadcastStore: broadcastStore,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		config: config,
	}
}

// CheckParallelizationRestriction checks if a new request can be created based on existing pending requests.
// If any pending request has a parallelization restriction (Parallelization > 0):
// 1. Only requests of the same type (same Action) can be added if the limit is not reached.
// 2. Requests with Parallelization 0 can always be added.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The new request to be checked.
func (r *Requestor) CheckParallelizationRestriction(
	ctx context.Context,
	session session.Session,
	request *callback.Request,
) bool {
	if request.Parallelization == 0 {
		return false
	}

	count := 0
	for _, rid := range session.GetNextRequests() {
		pr, prErr := r.requestorStore.GetRequest(ctx, rid)
		if prErr != nil || pr == nil {
			slog.WarnContext(ctx, "failed to retrieve pending request",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, rid),
				slog.Any(logging.FieldError, prErr),
			)
			continue
		}

		if pr.Parallelization != 0 {
			if pr.Action != request.Action {
				slog.WarnContext(ctx, "cannot create new request with different action",
					slog.String(logging.FieldSessionID, session.GetID()),
					slog.String(logging.FieldRequestID, rid),
					slog.String(logging.FieldAction, pr.Action),
					slog.String("requested_action", request.Action),
				)
				return true
			}
			count++
		}
	}

	if count >= request.Parallelization {
		slog.WarnContext(ctx, "parallelization limit reached",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Int("count", count),
			slog.String(logging.FieldAction, request.Action),
			slog.Int("parallelization", request.Parallelization),
		)
		return true
	}

	return false
}

// CreateRequest creates a new request and stores it in the requestor store. It also updates the session with the new request if necessary.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request to be sent.
// data: The data to be included in the request body.
func (r *Requestor) CreateRequest(
	ctx context.Context,
	session session.Session,
	request *callback.Request,
	data callback.RequestData,
) error {
	// FIXME: check session reloading especially when parallelization is used, as the session might be updated by other requests in the meantime
	if r.CheckParallelizationRestriction(ctx, session, request) {
		slog.WarnContext(ctx, "parallelization restriction violated",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.String(logging.FieldRequestID, request.ID),
		)
		return callback.ErrRequestNotAllowed
	}

	request.ID = uuid.NewString()
	data.SetCallback(fmt.Sprintf("%s/callback/%s", r.config.BaseURL, request.ID))

	body, rbErr := json.Marshal(data)
	if rbErr != nil {
		slog.ErrorContext(ctx, "failed to marshal request body",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, rbErr),
		)
		return callback.ErrInvalidRequest
	}
	request.Body = string(body)

	request.CreatedAt = time.Now().Unix()
	if uErr := r.requestorStore.CreateRequest(ctx, request); uErr != nil {
		slog.ErrorContext(ctx, "failed to store request",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, uErr),
		)
		return callback.ErrRequestNotSaved
	}

	if request.Parallelization != 0 {
		session.UpdateNextRequests(request.ID)
		if usErr := r.sessionStore.UpdateSession(ctx, session.GetID(), session); usErr != nil {
			slog.ErrorContext(ctx, "failed to update session with next request",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, request.ID),
				slog.Any(logging.FieldError, usErr),
			)
			return callback.ErrSessionNotUpdated
		}
	}

	return nil
}

// GetRequest retrieves a request from the cache based on its ID.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// requestID: The ID of the request to be retrieved.
func (r *Requestor) GetRequest(ctx context.Context, requestID string) (*callback.Request, error) {
	return r.requestorStore.GetRequest(ctx, requestID)
}

// CleanRequestQueue cleans the request queue for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session for which the request queue needs to be cleaned.
func (r *Requestor) CleanRequestQueue(ctx context.Context, session session.Session) error {
	if len(session.GetNextRequests()) == 0 {
		return nil
	}
	slog.InfoContext(ctx, "session has pending requests, cleaning up",
		slog.String(logging.FieldSessionID, session.GetID()),
	)

	rids := make([]string, len(session.GetNextRequests()))
	copy(rids, session.GetNextRequests())

	for _, rid := range rids {
		request, gErr := r.GetRequest(ctx, rid)
		if gErr != nil {
			slog.ErrorContext(ctx, "failed to retrieve request",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, rid),
				slog.Any(logging.FieldError, gErr),
			)
			return gErr
		}
		if request == nil {
			slog.WarnContext(ctx, "request not found, deleting from queue",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, rid),
			)
			session.DeleteNextRequest(rid)
			_ = r.sessionStore.UpdateSession(ctx, session.GetID(), session)
			continue
		}

		if err := r.requestorStore.CompleteRequest(ctx, session, request); err != nil {
			return err
		}
	}

	return nil
}

// Send sends requests to multiple URLs concurrently and handles their responses.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// urls: A slice of URLs to which the requests will be sent.
// requestBuilder: A function that builds a callback request based on the URL and session.
func (r *Requestor) Send(
	ctx context.Context,
	session session.Session,
	urls []string,
	requestBuilder RequestBuilderFunc,
) {
	bgCtx := context.WithoutCancel(ctx)

	for _, url := range urls {
		go func(u string) {
			request, data := requestBuilder(u)
			request.Endpoint = fmt.Sprintf("%s/%s", u, strings.ToLower(request.Action))

			unlock, lErr := r.sessionStore.LockSession(bgCtx, session.GetID())
			if lErr != nil {
				slog.ErrorContext(bgCtx, "failed to lock session",
					slog.String(logging.FieldSessionID, session.GetID()),
					slog.String(logging.FieldRequestID, request.ID),
					slog.Any(logging.FieldError, lErr),
				)
				// FIXME: what to do if we cannot lock the session? We cannot update the request status or broadcast an error message. Maybe we should panic here, as this is a critical error.
				return
			}
			defer unlock(bgCtx)

			rErr := r.CreateRequest(bgCtx, session, request, data)
			if rErr != nil {
				slog.ErrorContext(bgCtx, "failed to create request",
					slog.String(logging.FieldSessionID, session.GetID()),
					slog.String(logging.FieldEndpoint, u),
					slog.Any(logging.FieldError, rErr),
				)
				payload, _ := json.Marshal(
					fmt.Sprintf("failed to create request for session %s to %s: %v", session.GetID(), u, rErr),
				)
				_ = r.broadcastStore.Broadcast(bgCtx, session.GetID(), &message.Message{
					Event:   message.ErrorEvent,
					Payload: payload,
				})
				return
			}

			go r.superviseRequest(bgCtx, session, request)
		}(url)
	}
}

// superviseRequest supervises the lifecycle of a request, ensuring it is completed or handled appropriately.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request to be supervised.
func (r *Requestor) superviseRequest(ctx context.Context, session session.Session, request *callback.Request) {
	slog.DebugContext(ctx, "supervising request",
		slog.String(logging.FieldSessionID, session.GetID()),
		slog.String(logging.FieldRequestID, request.ID),
	)

	var errMsg string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		sErr := r.sendRequest(ctx, request)
		if sErr != nil {
			slog.WarnContext(ctx, "failed to send request",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, request.ID),
				slog.Int(logging.FieldAttempt, attempt+1),
				slog.Any(logging.FieldError, sErr),
			)

			if attempt < maxRetries {
				time.Sleep(randomRequestWaitTime())
			} else {
				slog.WarnContext(ctx, "max retries reached for sending request",
					slog.String(logging.FieldSessionID, session.GetID()),
					slog.String(logging.FieldRequestID, request.ID),
				)
				errMsg = fmt.Sprintf("could not send request %s to %s", request.ID, request.Endpoint)
			}

			continue
		}

		slog.DebugContext(ctx, "request sent successfully",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.String(logging.FieldRequestID, request.ID),
			slog.Int(logging.FieldAttempt, attempt+1),
		)

		cErr := r.waitForCallback(ctx, request)
		if cErr != nil {
			slog.WarnContext(ctx, "failed to receive callback",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, request.ID),
				slog.Int(logging.FieldAttempt, attempt+1),
				slog.Any(logging.FieldError, cErr),
			)

			if attempt >= maxRetries {
				slog.WarnContext(ctx, "max retries reached for receiving callback",
					slog.String(logging.FieldSessionID, session.GetID()),
					slog.String(logging.FieldRequestID, request.ID),
				)
				errMsg = fmt.Sprintf("could not receive callback for request %s to %s", request.ID, request.Endpoint)
			}

			continue
		}

		slog.DebugContext(ctx, "callback received",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.String(logging.FieldRequestID, request.ID),
			slog.Int(logging.FieldAttempt, attempt+1),
		)
		return
	}

	sessionID := session.GetID()
	unlock, lErr := r.sessionStore.LockSession(ctx, sessionID)
	if lErr != nil {
		slog.ErrorContext(ctx, "failed to lock session",
			slog.String(logging.FieldSessionID, sessionID),
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, lErr),
		)
		// FIXME: what to do if we cannot lock the session? We cannot update the request status or broadcast an error message. Maybe we should panic here, as this is a critical error.
		return
	}
	defer unlock(ctx)

	baseSession, sErr := r.sessionStore.GetBaseSession(ctx, sessionID)
	if sErr != nil || baseSession == nil {
		slog.ErrorContext(ctx, "failed to retrieve session",
			slog.String(logging.FieldSessionID, sessionID),
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, sErr),
		)
		errMsg = fmt.Sprintf("could not find session %s for request %s", sessionID, request.ID)
	} else {
		_ = r.requestorStore.CompleteRequest(ctx, baseSession, request)
	}

	payload, _ := json.Marshal(errMsg)
	_ = r.broadcastStore.Broadcast(ctx, sessionID, &message.Message{
		Event:   message.ErrorEvent,
		Payload: payload,
	})
}

// waitForCallback waits for a callback response for the given request. It checks the request status and handles timeouts or errors.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request for which to wait for a callback.
func (r *Requestor) waitForCallback(ctx context.Context, request *callback.Request) error {
	deadline := time.NewTimer(callbackTimeout)
	defer deadline.Stop()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			req, rErr := r.requestorStore.GetRequest(ctx, request.ID)
			if req == nil && rErr == nil {
				return nil
			}

		case <-deadline.C:
			return callback.ErrCallbackNotReceived

		case <-ctx.Done():
			return callback.ErrCallbackContextCancelled
		}
	}
}

// sendRequest sends a single request to the specified endpoint and handles the response.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// request: The request to be sent.
func (r *Requestor) sendRequest(
	ctx context.Context,
	request *callback.Request,
) error {
	req, crErr := http.NewRequestWithContext(ctx, http.MethodPost, request.Endpoint, strings.NewReader(request.Body))
	if crErr != nil {
		slog.ErrorContext(ctx, "failed to create HTTP request",
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, crErr),
		)
		return callback.ErrInvalidRequest
	}
	req.Header.Set("Content-Type", "application/json")

	resp, hErr := r.httpClient.Do(req)
	if hErr != nil {
		slog.ErrorContext(ctx, "failed to perform HTTP request",
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, hErr),
		)
		return callback.ErrTransportFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "request not accepted",
			slog.String(logging.FieldRequestID, request.ID),
			slog.Int("status_code", resp.StatusCode),
		)
		return callback.ErrNotAccepted
	}

	slog.DebugContext(ctx, "request sent successfully",
		slog.String(logging.FieldRequestID, request.ID),
	)
	return nil
}

// randomRequestWaitTime generates a random wait time between 1 second and maxRequestWaitTime for retrying requests.
//
//nolint:mnd // no magic numbers here, as this is a simple random wait time generator.
func randomRequestWaitTime() time.Duration {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxRequestWaitTime.Seconds())))
	if err != nil {
		nBig = big.NewInt(int64(maxRequestWaitTime.Seconds() / 2))
	}
	return time.Duration(1+nBig.Int64()) * time.Second
}
