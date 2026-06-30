package requestor

import (
	"context"
	"errors"
	"slices"

	"github.com/sirupsen/logrus"

	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/model"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/model"
	globalRequestor "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/requestor"
)

// Wrapper is a wrapper for the global requestor.
type Wrapper struct {
	requestor *globalRequestor.Requestor
	cache     *cache.Wrapper
}

// NewWrapper creates a new instance of the Wrapper.
// requestor: The global requestor instance to be wrapped.
func NewWrapper(requestor *globalRequestor.Requestor, cache *cache.Wrapper) *Wrapper {
	return &Wrapper{requestor: requestor, cache: cache}
}

// CanBeProcessed checks if a request is blocking and/or if it is the a request to be processed for the session.
// session: The current game session containing player connection information.
// request: The request to be checked.
func (w *Wrapper) CanBeProcessed(session *model.Session, request *callbackModel.Request) bool {
	return !request.Blocking || (request.Blocking && slices.Contains(session.NextRequests, request.ID))
}

// CheckSingletonBlockingRequest checks if any of the pending requests in the session are Singleton.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (w *Wrapper) CheckSingletonBlockingRequest(ctx context.Context, session *model.Session) bool {
	for _, rid := range session.NextRequests {
		pr, prErr := w.cache.GetRequest(ctx, rid)
		if prErr != nil {
			logrus.Errorf("failed to retrieve pending request %s for session %s: %v", rid, session.ID, prErr)
			return true
		}

		if pr.Singleton {
			logrus.Errorf("session %s has a Singleton pending request %s, cannot create new request", session.ID, rid)
			return true
		}
	}

	return false
}

// CreateRequest sends a request to the specified endpoint and stores it in the cache.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request to be sent.
// data: The data to be included in the request body.
func (w *Wrapper) CreateRequest(
	ctx context.Context,
	session *model.Session,
	request *callbackModel.Request,
	data any,
) error {
	if w.CheckSingletonBlockingRequest(ctx, session) {
		return errors.New("cannot create request while a Singleton request is pending")
	}

	rid := w.cache.GenerateUniqueSessionID(ctx)

	if rid != nil {
		request.ID = *rid
		rErr := w.requestor.CreateRequest(ctx, request, data)
		if rErr != nil {
			return rErr
		}

		if request.Blocking {
			session.NextRequests = append(session.NextRequests, request.ID)
			if err := w.cache.UpdateSession(ctx, session); err != nil {
				logrus.Errorf("failed to update session %s with next request %s: %v", session.ID, request.ID, err)
				return err
			}
		}

		return nil
	}

	logrus.Errorf("failed to generate unique request ID for session %s", session.ID)
	return errors.New("could not generate request")
}

// CompleteRequest marks a request as completed and updates the session accordingly.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request to be marked as completed.
func (w *Wrapper) CompleteRequest(ctx context.Context, session *model.Session, request *callbackModel.Request) error {
	cErr := w.requestor.CompleteRequest(ctx, request)
	if cErr != nil {
		logrus.Errorf("failed to update request %s for session %s: %v", request.ID, session.ID, cErr)
		return cErr
	}

	session.NextRequests = slices.DeleteFunc(session.NextRequests, func(id string) bool {
		return id == request.ID
	})
	if err := w.cache.UpdateSession(ctx, session); err != nil {
		logrus.Errorf("failed to update session %s after completing request %s: %v", session.ID, request.ID, err)
		return err
	}

	logrus.Debugf("request %s for session %s marked as completed", request.ID, session.ID)
	return nil
}

// CleanRequestQueue cleans the request queue for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session for which the request queue needs to be cleaned.
func (w *Wrapper) CleanRequestQueue(ctx context.Context, session *model.Session) error {
	if len(session.NextRequests) == 0 {
		return nil
	}
	logrus.Infof("session %s has pending requests, cleaning up", session.ID)

	for _, rid := range session.NextRequests {
		request, gErr := w.requestor.GetRequest(ctx, rid)
		if gErr != nil {
			logrus.Errorf("failed to retrieve request %s for session %s: %v", rid, session.ID, gErr)
			return gErr
		}

		if err := w.CompleteRequest(ctx, session, request); err != nil {
			return err
		}
	}

	return nil
}
