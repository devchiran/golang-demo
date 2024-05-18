package http

import (
	"golang-demo/internal"

	"github.com/gorilla/mux"
	"github.com/twitsprout/tools"
)

type Handler struct {
	Version    string
	router     *mux.Router
	Logger     tools.Logger
	AlbumStore internal.AlbumStore
}
