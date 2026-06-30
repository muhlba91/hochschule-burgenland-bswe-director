package requestor

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/model"
)

// httpTimeout defines the timeout duration for HTTP requests made by the Requestor.
const httpTimeout = 5 * time.Second

// FIXME: implement polling mechanism to check if the request is completed and retry

// Requestor is responsible for handling requests related to the Pokémon game flow.
type Requestor struct {
	cache      *cache.Cache
	httpClient *http.Client
}

// NewRequestor creates a new instance of Requestor.
func NewRequestor(cache *cache.Cache) *Requestor {
	return &Requestor{
		cache: cache,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// CreateRequest sends a request to the specified endpoint and stores it in the cache.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// request: The request to be sent.
// data: The data to be included in the request body.
func (r *Requestor) CreateRequest(
	ctx context.Context,
	request *callbackModel.Request,
	data any,
) error {
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
	if uErr := r.cache.UpdateRequest(ctx, request, cache.DefaultRequestExpiration); uErr != nil {
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
	request, err := r.cache.GetRequest(ctx, requestID)
	if err != nil {
		logrus.Errorf("failed to retrieve request %s: %v", requestID, err)
		return nil, err
	}
	return request, nil
}

// CompleteRequest marks a request as completed and updates the request accordingly.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// request: The request to be marked as completed.
func (r *Requestor) CompleteRequest(ctx context.Context, request *callbackModel.Request) error {
	request.Completed = true
	if err := r.cache.UpdateRequest(ctx, request, cache.DefaultRequestExpiration); err != nil {
		logrus.Errorf("failed to update request %s: %v", request.ID, err)
		return err
	}

	logrus.Debugf("request %s marked as completed", request.ID)
	return nil
}
