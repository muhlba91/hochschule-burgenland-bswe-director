package cache

import (
	"context"

	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/model"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/constants"
)

// GenerateUniqueRequestID generates a unique request ID for the given flow name.
// ctx: The context for the operation.
func (w *Wrapper) GenerateUniqueRequestID(ctx context.Context) *string {
	return w.cache.GenerateUniqueRequestID(ctx, constants.Name)
}

// CreateRequest creates a new request in the cache with the given request ID and data.
// ctx: The context for the operation.
// requestID: The unique request ID for the new request.
// data: The data to be stored in the request.
func (w *Wrapper) CreateRequest(ctx context.Context, data *callbackModel.Request) error {
	return w.cache.CreateRequest(ctx, data, requestExpirationTime)
}

// UpdateRequest updates an existing request in the cache with the given request ID and data.
// ctx: The context for the operation.
// requestID: The unique request ID for the existing request.
// data: The data to be stored in the request.
func (w *Wrapper) UpdateRequest(ctx context.Context, data *callbackModel.Request) error {
	return w.cache.UpdateRequest(ctx, data, requestExpirationTime)
}

// GetRequest retrieves the request data for the given request ID from the cache.
// ctx: The context for the operation.
// requestID: The unique request ID for the new request.
func (w *Wrapper) GetRequest(ctx context.Context, requestID string) (*callbackModel.Request, error) {
	return w.cache.GetRequest(ctx, requestID)
}
