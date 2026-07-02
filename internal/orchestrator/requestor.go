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
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: session.GetID(),
				logging.FieldRequestID: rid,
				logging.FieldError:     prErr,
			}).Warn("failed to retrieve pending request")
			continue
		}

		if pr.Parallelization != 0 {
			if pr.Action != request.Action {
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: session.GetID(),
					logging.FieldRequestID: rid,
					logging.FieldAction:    pr.Action,
					"requested_action":     request.Action,
				}).Warn("cannot create new request with different action")
				return true
			}
			count++
		}
	}

	if count >= request.Parallelization {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			"count":                count,
			logging.FieldAction:    request.Action,
			"parallelization":      request.Parallelization,
		}).Warn("parallelization limit reached")
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
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
		}).Warn("parallelization restriction violated")
		return callback.ErrRequestNotAllowed
	}

	request.ID = uuid.NewString()
	data.SetCallback(fmt.Sprintf("%s/callback/%s", r.config.BaseURL, request.ID))

	body, rbErr := json.Marshal(data)
	if rbErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
			logging.FieldError:     rbErr,
		}).Error("failed to marshal request body")
		return callback.ErrInvalidRequest
	}
	request.Body = string(body)

	request.CreatedAt = time.Now().Unix()
	if uErr := r.requestorStore.CreateRequest(ctx, request); uErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
			logging.FieldError:     uErr,
		}).Error("failed to store request")
		return callback.ErrRequestNotSaved
	}

	if request.Parallelization != 0 {
		session.UpdateNextRequests(request.ID)
		if usErr := r.sessionStore.UpdateSession(ctx, session.GetID(), session); usErr != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: session.GetID(),
				logging.FieldRequestID: request.ID,
				logging.FieldError:     usErr,
			}).Error("failed to update session with next request")
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
	logrus.WithFields(logrus.Fields{
		logging.FieldSessionID: session.GetID(),
	}).Info("session has pending requests, cleaning up")

	rids := make([]string, len(session.GetNextRequests()))
	copy(rids, session.GetNextRequests())

	for _, rid := range rids {
		request, gErr := r.GetRequest(ctx, rid)
		if gErr != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: session.GetID(),
				logging.FieldRequestID: rid,
				logging.FieldError:     gErr,
			}).Error("failed to retrieve request")
			return gErr
		}
		if request == nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: session.GetID(),
				logging.FieldRequestID: rid,
			}).Warn("request not found, deleting from queue")
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
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: session.GetID(),
					logging.FieldRequestID: request.ID,
					logging.FieldError:     lErr,
				}).Error("failed to lock session")
				// FIXME: what to do if we cannot lock the session? We cannot update the request status or broadcast an error message. Maybe we should panic here, as this is a critical error.
				return
			}
			defer unlock(bgCtx)

			rErr := r.CreateRequest(bgCtx, session, request, data)
			if rErr != nil {
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: session.GetID(),
					logging.FieldEndpoint:  u,
					logging.FieldError:     rErr,
				}).Error("failed to create request")
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
	logrus.WithFields(logrus.Fields{
		logging.FieldSessionID: session.GetID(),
		logging.FieldRequestID: request.ID,
	}).Debug("supervising request")

	var errMsg string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		sErr := r.sendRequest(ctx, request)
		if sErr != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: session.GetID(),
				logging.FieldRequestID: request.ID,
				logging.FieldAttempt:   attempt + 1,
				logging.FieldError:     sErr,
			}).Warn("failed to send request")

			if attempt < maxRetries {
				time.Sleep(randomRequestWaitTime())
			} else {
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: session.GetID(),
					logging.FieldRequestID: request.ID,
				}).Warn("max retries reached for sending request")
				errMsg = fmt.Sprintf("could not send request %s to %s", request.ID, request.Endpoint)
			}

			continue
		}

		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
			logging.FieldAttempt:   attempt + 1,
		}).Debug("request sent successfully")

		cErr := r.waitForCallback(ctx, request)
		if cErr != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: session.GetID(),
				logging.FieldRequestID: request.ID,
				logging.FieldAttempt:   attempt + 1,
				logging.FieldError:     cErr,
			}).Warn("failed to receive callback")

			if attempt >= maxRetries {
				logrus.WithFields(logrus.Fields{
					logging.FieldSessionID: session.GetID(),
					logging.FieldRequestID: request.ID,
				}).Warn("max retries reached for receiving callback")
				errMsg = fmt.Sprintf("could not receive callback for request %s to %s", request.ID, request.Endpoint)
			}

			continue
		}

		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
			logging.FieldAttempt:   attempt + 1,
		}).Debug("callback received")
		return
	}

	unlock, lErr := r.sessionStore.LockSession(ctx, session.GetID())
	if lErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
			logging.FieldError:     lErr,
		}).Error("failed to lock session")
		// FIXME: what to do if we cannot lock the session? We cannot update the request status or broadcast an error message. Maybe we should panic here, as this is a critical error.
		return
	}
	defer unlock(ctx)

	session, sErr := r.sessionStore.GetBaseSession(ctx, session.GetID())
	if sErr != nil || session == nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldRequestID: request.ID,
			logging.FieldError:     sErr,
		}).Error("failed to retrieve session")
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
		logrus.WithFields(logrus.Fields{
			logging.FieldRequestID: request.ID,
			logging.FieldError:     crErr,
		}).Error("failed to create HTTP request")
		return callback.ErrInvalidRequest
	}
	req.Header.Set("Content-Type", "application/json")

	resp, hErr := r.httpClient.Do(req)
	if hErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldRequestID: request.ID,
			logging.FieldError:     hErr,
		}).Error("failed to perform HTTP request")
		return callback.ErrTransportFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			logging.FieldRequestID: request.ID,
			"status_code":          resp.StatusCode,
		}).Error("request not accepted")
		return callback.ErrNotAccepted
	}

	logrus.WithFields(logrus.Fields{
		logging.FieldRequestID: request.ID,
	}).Debug("request sent successfully")
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
