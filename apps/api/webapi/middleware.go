package webapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/httpx/pbd"
	"github.com/samuelsih/nibiru/api/app"
)

type userContextKey struct{}

func (s Server) MiddlewareAuthenticated() httpx.Middleware {
	return func(next httpx.Handler) httpx.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			cookie, err := r.Cookie(s.sessionCookieName)
			if err != nil {
				return app.ErrInvalidSession
			}

			user, err := app.WhoAmI(r.Context(), s.db, cookie.Value)
			if err != nil {
				return err
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, user)

			return next(w, r.WithContext(ctx))
		}
	}
}

func MiddlewareLogger() httpx.Middleware {
	return func(next httpx.Handler) httpx.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			response := &responseRecorder{ResponseWriter: w}

			defer func(ctx context.Context) {
				recovered := recover()
				if recovered == http.ErrAbortHandler { //nolint:errorlint // compared by identity, as net/http does.
					panic(recovered)
				}

				if recovered != nil && response.status == 0 && r.Header.Get("Connection") != "Upgrade" {
					_ = pbd.New(http.StatusInternalServerError).Write(response)
				}

				status := response.status
				if status == 0 {
					status = http.StatusOK
				}

				duration := time.Since(start)

				level := slog.LevelInfo
				if recovered != nil || status >= http.StatusBadRequest {
					level = slog.LevelError
				}

				attrs := []slog.Attr{
					slog.String("http.request.method", r.Method),
					slog.String("url.path", r.URL.Path),
					slog.String("client.address", r.RemoteAddr),
					slog.String("server.address", r.Host),
					slog.String("network.protocol.version", r.Proto),
					slog.String("user_agent.original", r.UserAgent()),
					slog.String("http.request.id", httpx.RequestID(r)),
					slog.Int64("http.request.body.size", r.ContentLength),
					slog.Int("http.response.status_code", status),
					slog.Float64("http.server.request.duration", float64(duration)/float64(time.Millisecond)),
				}

				if recovered != nil {
					attrs = append(attrs,
						slog.String("error.message", "panic: "+fmt.Sprint(recovered)),
						slog.String("exception.stacktrace", string(debug.Stack())),
					)
				}

				slog.LogAttrs(ctx, level, "Request captured", attrs...)
			}(r.Context())

			return next(response, r)
		}
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return w.ResponseWriter.Write(b)
}

func (w *responseRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
