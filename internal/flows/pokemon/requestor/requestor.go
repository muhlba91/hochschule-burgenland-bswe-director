package requestor

import (
	"context"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
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
	return w.requestor.Send(ctx, session, urls, requestBuilder)
}
