package http

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/twitsprout/tools"
	"github.com/twitsprout/tools/requestid"
)

// NotFoundHandler returns an HTTP HandlerFunc that responds with the "Not Found"
// response.
func NotFoundHandler(logger tools.WarnLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONError(r.Context(), logger, w, r.URL.Query(), "http: not found", 404)
	})
}

// MethodNotAllowedHandler returns an HTTP HandlerFunc that responds with the
// "Method Not Allowed" response.
func MethodNotAllowedHandler(logger tools.WarnLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONError(r.Context(), logger, w, r.URL.Query(), "http: method not allowed", 405)
	})
}

// InternalServerErrorHandler returns an HTTP HandlerFunc that responds with the
// "Internal Server Error" response.
func InternalServerErrorHandler(logger tools.WarnLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONError(r.Context(), logger, w, r.URL.Query(), "http: internal server error", 500)
	})
}

// ServiceUnavailableHandler returns an HTTP HandlerFunc that responds with the
// "Service Unavailable" response.
func ServiceUnavailableHandler(logger tools.WarnLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONError(r.Context(), logger, w, r.URL.Query(), "http: service unavailable", 503)
	})
}

// VersionHandler returns an HTTP handler that responds with the standard JSON
// service response.
func VersionHandler(service, version string, logger tools.WarnLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type versionRes struct {
			Service string    `json:"service"`
			Time    time.Time `json:"time"`
			Version string    `json:"version"`
		}
		res := versionRes{
			Service: service,
			Time:    time.Now(),
			Version: version,
		}
		writeJSONData(r.Context(), logger, w, r.URL.Query(), res, 200)
	})
}

func writeJSONData(ctx context.Context, logger tools.WarnLogger, w http.ResponseWriter, v url.Values, data interface{}, code int) {
	err := WriteJSONData(w, v, data, code)
	if err != nil {
		logger.Warn("unable to write JSON data response",
			"request_id", requestid.Get(ctx),
			"details", err.Error(),
		)
	}
}

func writeJSONError(ctx context.Context, logger tools.WarnLogger, w http.ResponseWriter, v url.Values, msg string, code int) {
	err := WriteJSONError(w, v, msg, code)
	if err != nil {
		logger.Warn("unable to write JSON error response",
			"request_id", requestid.Get(ctx),
			"details", err.Error(),
		)
	}
}
