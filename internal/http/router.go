package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httputils "github.com/twitsprout/tools/http"
)

// Handler mounts all the handlers at the appropriate routes and adds any required middleware.
func (h *Handler) Handler() http.Handler {
	r := mux.NewRouter()

	r.Use(httputils.TimeoutMiddleware(1 * time.Minute))
	r.Use(httputils.RequestIDMiddleware)
	r.Use(httputils.RealIPMiddleware)
	r.Use(httputils.LimitReaderMiddleware(1 << 20))
	r.Use(httputils.LoggingMiddleware(h.Logger))
	r.Use(httputils.RecoverMiddleware(h.Logger, httputils.InternalServerErrorHandler(h.Logger)))
	r.Use(httputils.MaxConnectionsMiddleware(5000, httputils.ServiceUnavailableHandler(h.Logger)))
	r.Use(httputils.ConcurrentLimitMiddleware(250, httputils.ServiceUnavailableHandler(h.Logger)))

	r.MethodNotAllowedHandler = httputils.MethodNotAllowedHandler(h.Logger)
	r.NotFoundHandler = httputils.NotFoundHandler(h.Logger)

	versionHandler := httputils.VersionHandler(h.AppName, h.Version, h.Logger)
	r.Methods("GET").Path("/").Name("root").Handler(versionHandler)
	r.Methods("GET").Path("/version").Name("version").Handler(versionHandler)

	v1 := r.PathPrefix("/v1").Subrouter()

	v1.Methods("GET").Path("/albums").Name("list_albums").HandlerFunc(h.ListAlbums)
	v1.Methods("GET").Path("/album/{id}").Name("get_album").HandlerFunc(h.GetAlbum)
	v1.Methods("POST").Path("/album").Name("create_album").HandlerFunc(h.CreateAlbum)
	h.router = r
	return r
}
