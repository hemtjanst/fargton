package bridge

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "färgton/middleware context value " + k.name
}

var (
	// AuthenticatedCtxKey is the context.Context key to store the request log entry.
	AuthenticatedCtxKey = &contextKey{"Authenticated"}
)

// Authenticate sets a key to indicate whether the request came from a
// whitelisted user or not
func (s *Server) Authenticate(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := chi.RouteContext(ctx).URLParam("userID")

		auth := s.config.authDisabled
		s.config.RLock()
		wt := *s.config.Whitelist
		if _, ok := wt[userID]; ok {
			auth = true
		}
		s.config.RUnlock()

		ctx = context.WithValue(ctx, AuthenticatedCtxKey, auth)
		r = r.WithContext(ctx)

		if !auth && infoFromRequest(r).resource != "/config" {
			renderListOK(w, r, errUnauthorized(r))
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
