package requestor

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
	pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// Wrapper is a wrapper for the global requestor.
type Wrapper struct {
	requestor *orchestrator.Requestor
	store     *store.Wrapper
}

// NewWrapper creates a new instance of the Wrapper.
// requestor: The global requestor instance to be wrapped.
func NewWrapper(requestor *orchestrator.Requestor, store *store.Wrapper) *Wrapper {
	return &Wrapper{requestor: requestor, store: store}
}

// CleanRequestQueue cleans the request queue for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session for which the request queue needs to be cleaned.
func (w *Wrapper) CleanRequestQueue(ctx context.Context, session *session.Session) error {
	return w.requestor.CleanRequestQueue(ctx, session)
}

// Send sends requests to multiple URLs concurrently and handles their responses.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// urls: A slice of URLs to which the requests will be sent.
// requestBuilder: A function that builds a callback request based on the URL and session.
func (w *Wrapper) Send(
	ctx context.Context,
	session *session.Session,
	urls []string,
	requestBuilder orchestrator.RequestBuilderFunc,
) error {
	return w.requestor.Send(ctx, session, urls, requestBuilder, w.createRequest)
}

// createRequest sends a request to the specified endpoint and stores it in the cache.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// request: The request to be sent.
// data: The data to be included in the request body.
func (w *Wrapper) createRequest(
	ctx context.Context,
	session pkgSession.Session,
	request *callbackModel.Request,
	data callbackModel.RequestData,
) error {
	rErr := w.requestor.CreateRequest(ctx, session, request, data)
	if rErr != nil {
		return rErr
	}

	if request.Parallelization != 0 {
		session.UpdateNextRequests(request.ID)
		if err := w.store.UpdateSession(ctx, session); err != nil {
			logrus.Errorf("failed to update session %s with next request %s: %v", session.GetID(), request.ID, err)
			return err
		}

		return nil
	}

	logrus.Errorf("failed to generate unique request ID for session %s", session.GetID())
	return errors.New("could not generate request")
}
