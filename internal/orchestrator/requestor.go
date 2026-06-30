package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// httpTimeout defines the timeout duration for HTTP requests made by the Requestor.
const httpTimeout = 5 * time.Second

// RequestBuilderFunc is a function type that takes a session ID and a session object, and returns a callback request.
type RequestBuilderFunc func() (*callbackModel.Request, callbackModel.RequestData)

// Requestor is responsible for handling requests related to the Pokémon game flow.
type Requestor struct {
	sessionStore   store.SessionStore
	requestorStore store.RequestStore
	httpClient     *http.Client
}

// NewRequestor creates a new instance of Requestor.
// requestorStore: The store for managing requests.
// sessionStore: The store for managing sessions.
func NewRequestor(requestorStore store.RequestStore, sessionStore store.SessionStore) *Requestor {
	return &Requestor{
		sessionStore:   sessionStore,
		requestorStore: requestorStore,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
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
	request *callbackModel.Request,
) bool {
	if request.Parallelization == 0 {
		return false
	}

	count := 0
	for _, rid := range session.GetNextRequests() {
		pr, prErr := r.requestorStore.GetRequest(ctx, rid)
		if prErr != nil {
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

// CreateRequest sends a request to the specified endpoint and stores it in the cache.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request to be sent.
// data: The data to be included in the request body.
func (r *Requestor) CreateRequest(
	ctx context.Context,
	session session.Session,
	request *callbackModel.Request,
	data callbackModel.RequestData,
) error {
	if r.CheckParallelizationRestriction(ctx, session, request) {
		logrus.Errorf("parallelization restriction violated for request %s", request.ID)
		return fmt.Errorf("parallelization restriction violated for request %s", request.ID)
	}

	request.ID = uuid.NewString()
	// FIXME: get base url from config or environment variable
	data.SetCallback(fmt.Sprintf("%s/callback/%s", "<base_url>", request.ID))

	body, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("failed to marshal request body for request %s: %v", request.ID, err)
		return err
	}
	request.Body = string(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, request.Endpoint, strings.NewReader(request.Body))
	if err != nil {
		logrus.Errorf("failed to create request for request %s: %v", request.ID, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		logrus.Errorf("failed to send request for request %s: %v", request.ID, err)
		return err
	}
	defer resp.Body.Close()

	request.CreatedAt = time.Now().Unix()
	if uErr := r.requestorStore.UpdateRequest(ctx, request); uErr != nil {
		logrus.Errorf("failed to store request for request %s: %v", request.ID, uErr)
		return uErr
	}
	// FIXME: refactor to handle different status codes and retry logic

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		logrus.Errorf("request for request %s returned status %d", request.ID, resp.StatusCode)
		return err
	}

	logrus.Debugf("request for request %s sent successfully", request.ID)
	return nil
}

// GetRequest retrieves a request from the cache based on its ID.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// requestID: The ID of the request to be retrieved.
func (r *Requestor) GetRequest(ctx context.Context, requestID string) (*callbackModel.Request, error) {
	request, err := r.requestorStore.GetRequest(ctx, requestID)
	if err != nil {
		logrus.Errorf("failed to retrieve request %s: %v", requestID, err)
		return nil, err
	}
	return request, nil
}

// CompleteRequest marks a request as completed and updates the request accordingly.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The session to which the request belongs.
// request: The request to be marked as completed.
func (r *Requestor) CompleteRequest(
	ctx context.Context,
	session session.Session,
	request *callbackModel.Request,
) error {
	request.Completed = true
	if err := r.requestorStore.UpdateRequest(ctx, request); err != nil {
		logrus.Errorf("failed to update request %s: %v", request.ID, err)
		return err
	}

	session.DeleteNextRequest(request.ID)
	if err := r.sessionStore.UpdateSession(ctx, session.GetID(), session); err != nil {
		logrus.Errorf("failed to update session %s after completing request %s: %v", session.GetID(), request.ID, err)
		return err
	}

	logrus.Debugf("request %s marked as completed", request.ID)
	return nil
}

// CleanRequestQueue cleans the request queue for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session for which the request queue needs to be cleaned.
func (r *Requestor) CleanRequestQueue(ctx context.Context, session session.Session) error {
	if len(session.GetNextRequests()) == 0 {
		return nil
	}
	logrus.Infof("session %s has pending requests, cleaning up", session.GetID())

	for _, rid := range session.GetNextRequests() {
		request, gErr := r.GetRequest(ctx, rid)
		if gErr != nil {
			logrus.Errorf("failed to retrieve request %s for session %s: %v", rid, session.GetID(), gErr)
			return gErr
		}

		if err := r.CompleteRequest(ctx, session, request); err != nil {
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
// createRequestFunction: A function that creates a new request.
func (r *Requestor) Send(
	ctx context.Context,
	session session.Session,
	urls []string,
	requestBuilder RequestBuilderFunc,
	createRequestFunction func(context.Context, session.Session, *callbackModel.Request, callbackModel.RequestData) error,
) error {
	// FIXME: implement polling mechanism to check if the request is completed and retry

	var wg sync.WaitGroup
	errCh := make(chan error, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			request, data := requestBuilder()
			request.Endpoint = fmt.Sprintf("%s/%s", url, strings.ToLower(request.Action))

			rErr := createRequestFunction(ctx, session, request, data)
			if rErr != nil {
				errCh <- fmt.Errorf("failed to create request for player at %s: %w", u, rErr)
			}
		}(url)
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered errors while sending requests for session %s", session.GetID())
	}

	logrus.Debugf("all requests responded successfully for session %s", session.GetID())
	return nil
}
