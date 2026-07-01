package redis

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// CreateRequest creates a new request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
func (c *Cache) CreateRequest(ctx context.Context, request *callback.Request) error {
	if err := c.Set(ctx, request.ID, request, constants.DefaultRequestExpiration); err != nil {
		logrus.Errorf("failed to save request: %v", err)
		return err
	}

	return nil
}

// UpdateRequest updates an existing request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
func (c *Cache) UpdateRequest(ctx context.Context, request *callback.Request) error {
	logrus.Debugf("updating request with ID: %s", request.ID)
	return c.CreateRequest(ctx, request)
}

// GetRequest retrieves the request data for the given request ID from the cache.
// ctx: The context for the operation.
// requestID: The unique request ID for the new request.
func (c *Cache) GetRequest(ctx context.Context, requestID string) (*callback.Request, error) {
	request, err := c.Get(ctx, requestID)
	if err != nil || request == nil {
		return nil, err
	}

	var requestModel callback.Request
	if uErr := json.Unmarshal([]byte(*request), &requestModel); uErr != nil {
		logrus.Errorf("failed to unmarshal request data: %v", uErr)
		return nil, uErr
	}

	return &requestModel, nil
}

// CompleteRequest marks a request as completed and updates the request accordingly.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The session to which the request belongs.
// request: The request to be marked as completed.
func (c *Cache) CompleteRequest(
	ctx context.Context,
	session session.Session,
	request *callback.Request,
) error {
	session.DeleteNextRequest(request.ID)
	if err := c.UpdateSession(ctx, session.GetID(), session); err != nil {
		logrus.Errorf("failed to update session %s after completing request %s: %v", session.GetID(), request.ID, err)
		return err
	}

	if err := c.Delete(ctx, request.ID); err != nil {
		logrus.Errorf("failed to delete request: %v", err)
		return err
	}

	logrus.Debugf("request %s marked as completed and deleted", request.ID)
	return nil
}
