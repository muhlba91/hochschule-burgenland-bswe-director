package redis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// CreateRequest creates a new request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
func (c *Cache) CreateRequest(ctx context.Context, request *callback.Request) error {
	requestData, _ := json.Marshal(request)

	logrus.Debugf("saving request with ID: %s, data: %s", request.ID, requestData)
	if err := c.client.Set(ctx, request.ID, requestData, constants.DefaultRequestExpiration).Err(); err != nil {
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
//
//nolint:nilnil // This function returns nil, nil when the request is not found, which is a valid case.
func (c *Cache) GetRequest(ctx context.Context, requestID string) (*callback.Request, error) {
	request, err := c.client.Get(ctx, requestID).Result()
	logrus.Debugf("retrieved request data for key %s: %s", requestID, request)

	if errors.Is(err, redis.Nil) {
		logrus.Warnf("request not found: %v", err)
		return nil, nil
	} else if err != nil {
		logrus.Errorf("failed to get request: %v", err)
		return nil, err
	}

	var requestModel callback.Request
	if uErr := json.Unmarshal([]byte(request), &requestModel); uErr != nil {
		logrus.Errorf("failed to unmarshal request data: %v", uErr)
		return nil, uErr
	}

	return &requestModel, nil
}
