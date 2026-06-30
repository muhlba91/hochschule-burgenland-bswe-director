package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/model"
)

// GenerateUniqueRequestID generates a unique request ID for the given flow name.
// ctx: The context for the operation.
// flowName: The name of the flow for which to generate the request ID.
func (c *Cache) GenerateUniqueRequestID(ctx context.Context, flowName string) *string {
	var id string

	for range maxIterations {
		id = uuid.NewString()
		key := fmt.Sprintf("request:%s:%s", flowName, id)

		exists, err := c.Exists(ctx, key)
		if err != nil {
			logrus.WithContext(ctx).Errorf("failed to check if request exists: %v", err)
			continue
		}
		if exists == 0 {
			return &key
		}
	}

	return nil
}

// CreateRequest creates a new request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
// expiration: The expiration time for the request.
func (c *Cache) CreateRequest(ctx context.Context, request *model.Request, expiration time.Duration) error {
	requestData, _ := json.Marshal(request)

	logrus.Debugf("saving request with ID: %s, data: %s", request.ID, requestData)
	if err := c.client.Set(ctx, request.ID, requestData, expiration).Err(); err != nil {
		logrus.Errorf("failed to save request: %v", err)
		return err
	}

	return nil
}

// UpdateRequest updates an existing request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
// expiration: The expiration time for the request.
func (c *Cache) UpdateRequest(ctx context.Context, request *model.Request, expiration time.Duration) error {
	logrus.Debugf("updating request with ID: %s", request.ID)
	return c.CreateRequest(ctx, request, expiration)
}

// GetRequest retrieves the request data for the given request ID from the cache.
// ctx: The context for the operation.
// requestID: The unique request ID for the new request.
//
//nolint:nilnil // This function returns nil, nil when the request is not found, which is a valid case.
func (c *Cache) GetRequest(ctx context.Context, requestID string) (*model.Request, error) {
	request, err := c.client.Get(ctx, requestID).Result()
	logrus.Debugf("retrieved request data for key %s: %s", requestID, request)

	if errors.Is(err, redis.Nil) {
		logrus.Warnf("request not found: %v", err)
		return nil, nil
	} else if err != nil {
		logrus.Errorf("failed to get request: %v", err)
		return nil, err
	}

	var requestModel model.Request
	if uErr := json.Unmarshal([]byte(request), &requestModel); uErr != nil {
		logrus.Errorf("failed to unmarshal request data: %v", uErr)
		return nil, uErr
	}

	return &requestModel, nil
}
