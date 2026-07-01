package orchestrator

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
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
			logrus.Errorf("failed to retrieve pending request %s for session %s: %v", rid, session.GetID(), prErr)
			continue
		}

		if pr.Parallelization != 0 {
			if pr.Action != request.Action {
				logrus.Errorf(
					"session %s has a pending request %s with action %s, cannot create new request with action %s",
					session.GetID(),
					rid,
					pr.Action,
					request.Action,
				)
				return true
			}
			count++
		}
	}

	if count >= request.Parallelization {
		logrus.Errorf(
			"session %s already has %d pending requests of type %s, parallelization limit %d reached",
			session.GetID(),
			count,
			request.Action,
			request.Parallelization,
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
	if r.CheckParallelizationRestriction(ctx, session, request) {
		logrus.Errorf("parallelization restriction violated for request %s", request.ID)
		return callback.ErrRequestNotAllowed
	}

	request.ID = uuid.NewString()
	data.SetCallback(fmt.Sprintf("%s/callback/%s", r.config.BaseURL, request.ID))

	body, rbErr := json.Marshal(data)
	if rbErr != nil {
		logrus.Errorf("failed to marshal request body for request %s: %v", request.ID, rbErr)
		return callback.ErrInvalidRequest
	}
	request.Body = string(body)

	request.CreatedAt = time.Now().Unix()
	if uErr := r.requestorStore.CreateRequest(ctx, request); uErr != nil {
		logrus.Errorf("failed to store request for request %s: %v", request.ID, uErr)
		return callback.ErrRequestNotSaved
	}

	if request.Parallelization != 0 {
		session.UpdateNextRequests(request.ID)
		if usErr := r.sessionStore.UpdateSession(ctx, session.GetID(), session); usErr != nil {
			logrus.Errorf("failed to update session %s with next request %s: %v", session.GetID(), request.ID, usErr)
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
	logrus.Infof("session %s has pending requests, cleaning up", session.GetID())

	rids := make([]string, len(session.GetNextRequests()))
	copy(rids, session.GetNextRequests())

	for _, rid := range rids {
		request, gErr := r.GetRequest(ctx, rid)
		if gErr != nil {
			logrus.Errorf("failed to retrieve request %s for session %s: %v", rid, session.GetID(), gErr)
			return gErr
		}
		if request == nil {
			logrus.Warnf("request %s not found for session %s, deleting from queue", rid, session.GetID())
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

			rErr := r.CreateRequest(bgCtx, session, request, data)
			if rErr != nil {
				logrus.Errorf("failed to create request for session %s to %s: %v", session.GetID(), u, rErr)
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
	logrus.Debugf("supervising request %s for session %s", request.ID, session.GetID())

	var errMsg string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		sErr := r.sendRequest(ctx, request)
		if sErr != nil {
			logrus.Warnf("attempt %d: failed to send request %s: %v", attempt+1, request.ID, sErr)

			if attempt < maxRetries {
				time.Sleep(randomRequestWaitTime())
			} else {
				logrus.Errorf("max retries reached for request %s", request.ID)
				errMsg = fmt.Sprintf("could not send request %s to %s", request.ID, request.Endpoint)
			}

			continue
		}

		logrus.Debugf("request %s sent successfully for attempt %d", request.ID, attempt+1)

		cErr := r.waitForCallback(ctx, request)
		if cErr != nil {
			logrus.Warnf("attempt %d: failed to receive callback for request %s: %v", attempt+1, request.ID, cErr)

			if attempt >= maxRetries {
				logrus.Errorf("max retries reached for request %s", request.ID)
				errMsg = fmt.Sprintf("could not receive callback for request %s to %s", request.ID, request.Endpoint)
			}

			continue
		}

		logrus.Debugf("callback received for request %s on attempt %d", request.ID, attempt+1)
		return
	}

	unlock, lErr := r.sessionStore.LockSession(ctx, session.GetID())
	if lErr != nil {
		logrus.Errorf("failed to lock session %s for request %s: %v", session.GetID(), request.ID, lErr)
		return
	}
	defer unlock(ctx)

	session, sErr := r.sessionStore.GetBaseSession(ctx, session.GetID())
	if sErr != nil || session == nil {
		logrus.Errorf("failed to retrieve session %s for request %s: %v", session.GetID(), request.ID, sErr)
		errMsg = fmt.Sprintf("could not find session %s for request %s", session.GetID(), request.ID)
	} else {
		_ = r.requestorStore.CompleteRequest(ctx, session, request)
	}

	payload, _ := json.Marshal(errMsg)
	_ = r.broadcastStore.Broadcast(ctx, session.GetID(), &message.Message{
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
		logrus.Errorf("failed to create request for request %s: %v", request.ID, crErr)
		return callback.ErrInvalidRequest
	}
	req.Header.Set("Content-Type", "application/json")

	resp, hErr := r.httpClient.Do(req)
	if hErr != nil {
		logrus.Errorf("failed to send request for request %s: %v", request.ID, hErr)
		return callback.ErrTransportFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		logrus.Errorf("request for request %s returned status %d", request.ID, resp.StatusCode)
		return callback.ErrNotAccepted
	}

	logrus.Debugf("request for request %s sent successfully", request.ID)
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
