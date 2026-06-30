package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// GenerateUniqueRequestID generates a unique request ID for the given flow name.
// ctx: The context for the operation.
func (w *Wrapper) GenerateUniqueRequestID() string {
	return callback.GetRequestID(constants.Name, uuid.NewString())
}

// CreateRequest creates a new request in the cache with the given request ID and data.
// ctx: The context for the operation.
// requestID: The unique request ID for the new request.
// data: The data to be stored in the request.
func (w *Wrapper) CreateRequest(ctx context.Context, data *callback.Request) error {
	return w.requestStore.CreateRequest(ctx, data)
}

// UpdateRequest updates an existing request in the cache with the given request ID and data.
// ctx: The context for the operation.
// requestID: The unique request ID for the existing request.
// data: The data to be stored in the request.
func (w *Wrapper) UpdateRequest(ctx context.Context, data *callback.Request) error {
	return w.requestStore.UpdateRequest(ctx, data)
}

// GetRequest retrieves the request data for the given request ID from the cache.
// ctx: The context for the operation.
// requestID: The unique request ID for the new request.
func (w *Wrapper) GetRequest(ctx context.Context, requestID string) (*callback.Request, error) {
	return w.requestStore.GetRequest(ctx, requestID)
}
