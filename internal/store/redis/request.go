package redis

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// CreateRequest creates a new request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
func (c *Cache) CreateRequest(ctx context.Context, request *callback.Request) error {
	if err := c.Set(ctx, request.ID, request, constants.DefaultRequestExpiration); err != nil {
		slog.ErrorContext(ctx, "failed to save request",
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, err),
		)
		return err
	}

	return nil
}

// UpdateRequest updates an existing request in the cache with the given request ID and data.
// ctx: The context for the operation.
// request: The request data to be stored in the cache.
func (c *Cache) UpdateRequest(ctx context.Context, request *callback.Request) error {
	slog.DebugContext(ctx, "updating request",
		slog.String(logging.FieldRequestID, request.ID),
	)
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
		slog.ErrorContext(ctx, "failed to unmarshal request data",
			slog.String(logging.FieldRequestID, requestID),
			slog.Any(logging.FieldError, uErr),
		)
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
	if request.Parallelization != 0 {
		session.DeleteNextRequest(request.ID)
		if err := c.UpdateSession(ctx, session.GetID(), session); err != nil {
			slog.ErrorContext(ctx, "failed to update session for deleting the request",
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, request.ID),
				slog.Any(logging.FieldError, err),
			)
			return err
		}
	}

	if err := c.Delete(ctx, request.ID); err != nil {
		slog.ErrorContext(ctx, "failed to delete request",
			slog.String(logging.FieldRequestID, request.ID),
			slog.Any(logging.FieldError, err),
		)
		return err
	}

	slog.DebugContext(ctx, "deleted request",
		slog.String(logging.FieldRequestID, request.ID),
	)
	return nil
}
