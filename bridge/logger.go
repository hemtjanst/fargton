package bridge

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return.
func Logger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				info := infoFromRequest(r)
				l.Info("Served",
					zap.String("method", info.method),
					zap.Int("status", ww.Status()),
					zap.String("path", info.resource),
					zap.String("userID", info.uid),
					zap.String("proto", info.httpVersion),
					zap.Bool("tls", info.tls),
					zap.String("source", info.sourceIP),
					zap.Duration("elapsed", time.Since(t1)),
					zap.Int("size", ww.BytesWritten()),
					zap.String("requestID", middleware.GetReqID(r.Context())),
				)
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
