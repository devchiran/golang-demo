package http

import (
	"errors"
	cl "golang-demo/pkg/catelog"
	"net/http"

	"github.com/ryanfowler/uuid"
	httputils "github.com/twitsprout/tools/http"
	"github.com/twitsprout/tools/requestid"
)

// ListAlbums get the list of all the albums
func (h *Handler) ListAlbums(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	v := r.URL.Query()
	reqID := requestid.Get(ctx)

	res, err := h.AlbumStore.ListAlbums(ctx)
	if err != nil {
		if err == cl.ErrNotFound {
			h.Logger.Error("[ListAlbums] No albums found",
				"request_id", reqID,
				"details", err.Error(),
			)
			_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusNotFound)
			return
		}

		h.Logger.Error("[ListAlbums] error getting albums list",
			"request_id", reqID,
			"details", err.Error(),
		)
		_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = httputils.WriteJSON(w, v, res, http.StatusOK)
}

// GetAlbum get the details of a album matcing with the album_id
func (h *Handler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	v := r.URL.Query()
	reqID := requestid.Get(ctx)

	req, err := parseGetAlbumRequest(r)
	if err != nil {
		h.Logger.Error("[GetAlbum] error parsing request",
			"request_id", reqID,
			"details", err.Error())
		_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.AlbumStore.GetAlbum(ctx, req.AlbumID)
	if err != nil {
		if err == cl.ErrNotFound {
			h.Logger.Error("[GetAlbum] no album found",
				"request_id", reqID,
				"details", err.Error(),
			)
			_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusNotFound)
			return
		}

		h.Logger.Error("[GetAlbum] error getting album",
			"request_id", reqID,
			"details", err.Error(),
		)
		_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = httputils.WriteJSON(w, v, res, http.StatusOK)
}

func parseGetAlbumRequest(r *http.Request) (cl.GetAlbumReq, error) {
	var req cl.GetAlbumReq
	v := r.URL.Query()

	albumID := v.Get("id")
	if albumID == "-" || albumID == "" {
		return req, errors.New("[parseGetAlbumRequest] album id must be provided")
	}

	req = cl.GetAlbumReq{
		AlbumID: albumID,
	}
	return req, nil
}

// CreateAlbum creates a album with the requested title
func (h *Handler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	v := r.URL.Query()
	reqID := requestid.Get(ctx)

	req, err := parseCreateAlbumRequest(r)
	if err != nil {
		h.Logger.Error("[CreateReview] error parsing request",
			"request_id", reqID,
			"details", err.Error())
		_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.AlbumStore.CreateAlbum(ctx, req)
	if err != nil {
		h.Logger.Error("[CreateReview] error storing review",
			"request_id", reqID,
			"details", err.Error(),
		)
		_ = httputils.WriteJSONError(w, v, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = httputils.WriteJSON(w, v, res, http.StatusCreated)
}

func parseCreateAlbumRequest(r *http.Request) (cl.CreateAlbumRequest, error) {
	var req cl.CreateAlbumRequest
	v := r.URL.Query()

	// Generate album id to add to database
	albumID, err := uuid.NewV4()
	if err != nil {
		return req, errors.New("[parseGetAlbumRequest] album id must be provided")
	}

	albumTitle := v.Get("title")
	if albumTitle == "" || albumTitle == " " {
		return req, errors.New("[parseGetAlbumRequest] album id must be provided")
	}

	req = cl.CreateAlbumRequest{
		AlbumID: albumID.String(),
		Title:   albumTitle,
	}
	return req, nil
}
