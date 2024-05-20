package http

import (
	"context"
	"golang-demo/internal/mock"
	cl "golang-demo/pkg/catelog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	httputils "github.com/twitsprout/tools/http"
	jsonutils "github.com/twitsprout/tools/json"
	tm "github.com/twitsprout/tools/mock"
)

func TestCreateAlbum(t *testing.T) {
	album := cl.Album{
		ID:        "1234",
		Title:     "Mountains",
		CreatedAt: time.Time{},
		UpdatedAt: time.Now(),
	}
	albumWithoutTitle := cl.Album{
		ID: "1234",
	}
	url := "/v1/album"
	table := []struct {
		label         string
		url           string
		body          string
		createAlbumFn func(ctx context.Context, r cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error)
		expCode       int
		expRes        interface{}
	}{
		{
			label:   "should fail if there's an error decoding json",
			url:     url,
			body:    `{badjson`,
			expCode: http.StatusBadRequest,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "json: invalid character 'b' looking for beginning of object key string: '{badjson'",
				},
			},
		},
		{
			label: "should fail if album id is missing",
			url:   url,
			body: `{
				"title": "Test album",
				"created_at": "2024-05-06T20:11:04.272642Z",
				"updated_at": "2024-05-06T20:11:04.272642Z"
			}`,
			expCode: http.StatusBadRequest,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "[parseCreateAlbumRequest] Album id must be provided",
				},
			},
		},
		{
			label: "should fail if album title is missing",
			url:   url,
			body: `{
				"id": "1234",
				"created_at": "2024-05-06T20:11:04.272642Z",
				"updated_at": "2024-05-06T20:11:04.272642Z"
			}`,
			createAlbumFn: func(ctx context.Context, r cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error) {
				return cl.CreateAlbumResponse{
					Album: &albumWithoutTitle,
				}, nil
			},
			expCode: http.StatusCreated,
			expRes: cl.CreateAlbumResponse{
				Album: &albumWithoutTitle,
			},
		},
		{
			label: "should fail if createReviewFn fails",
			url:   url,
			body: `{
				"id": "1234",
				"title": "Text title",
				"created_at": "2024-05-06T20:11:04.272642Z",
				"updated_at": "2024-05-06T20:11:04.272642Z"
			}`,
			createAlbumFn: func(ctx context.Context, r cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error) {
				return cl.CreateAlbumResponse{}, errors.New("internal server error")
			},
			expCode: http.StatusInternalServerError,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "internal server error",
				},
			},
		},
		{
			label: "should pass with all valid fields",
			url:   url,
			body: `{
				"id": "1234",
				"title": "Text title",
				"created_at": "2024-05-06T20:11:04.272642Z",
				"updated_at": "2024-05-06T20:11:04.272642Z"
			}`,
			createAlbumFn: func(ctx context.Context, r cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error) {
				return cl.CreateAlbumResponse{
					Album: &album,
				}, nil
			},
			expCode: http.StatusCreated,
			expRes: cl.CreateAlbumResponse{
				Album: &album,
			},
		},
	}
	for i := 0; i < len(table); i++ {
		ts := table[i]
		t.Run(ts.label, func(t *testing.T) {
			h := Handler{
				AlbumStore: &mock.AlbumStore{
					CreateAlbumFn: ts.CreateAlbumFn,
				},
				Logger: tm.NopLogger,
			}

			h.Handler()
			wr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", ts.url, strings.NewReader(ts.body))
			h.router.ServeHTTP(wr, req)

			if wr.Code != ts.expCode {
				var res httputils.JSONErrRes
				err := jsonutils.Decode(wr.Body, &res)
				if err != nil {
					t.Fatalf("unexpected error returned from decoding response body: %s", err.Error())
				}

				t.Fatalf("unexpected response code returned: %s %s", cmp.Diff(ts.expCode, wr.Code), res.Error.Message)
			}

			if wr.Code != 201 {
				var res httputils.JSONErrRes
				err := jsonutils.Decode(wr.Body, &res)
				if err != nil {
					t.Fatalf("unexpected error returned from decoding response body: %s", err.Error())
				}

				if !cmp.Equal(res, ts.expRes) {
					t.Fatalf("unexpected response returned: %s", cmp.Diff(res, ts.expRes))
				}
			} else {
				var res cl.GetAlbumRes
				err := jsonutils.Decode(wr.Body, &res)
				if err != nil {
					t.Fatalf("unexpected error returned from decoding response body: %s", err.Error())
				}

				if !cmp.Equal(res, ts.expRes) {
					t.Fatalf("unexpected response returned: %s", cmp.Diff(res, ts.expRes))
				}
			}
		})
	}
}

func TestGetAlbum(t *testing.T) {
	album := cl.Album{
		ID:        "1234",
		Title:     "test",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	url := "/v1/album"
	table := []struct {
		label      string
		url        string
		getAlbumFn func(ctx context.Context, id string) (cl.GetAlbumRes, error)
		expCode    int
		expRes     interface{}
	}{
		{
			label:   "should fail if there's no album id provided",
			url:     url,
			expCode: http.StatusBadRequest,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "[parseGetAlbumRequest] album id must be provided",
				},
			},
		},
		{
			label: "should fail if there's a dash album id provided",
			url:   url + "/",

			expCode: http.StatusBadRequest,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "[parseGetAlbumRequest] album id must be provided",
				},
			},
		},
		{
			label: "should fail if getAlbumFn fails",
			url:   url + "/1234",
			getAlbumFn: func(ctx context.Context, id string) (cl.GetAlbumRes, error) {
				return cl.GetAlbumRes{}, errors.New("internal server error")
			},
			expCode: http.StatusInternalServerError,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "internal server error",
				},
			},
		},
		{
			label: "should fail if getAlbumFn finds no rows",
			url:   url + "/9999",
			getAlbumFn: func(ctx context.Context, id string) (cl.GetAlbumRes, error) {
				return cl.GetAlbumRes{}, cl.ErrNotFound
			},
			expCode: http.StatusNotFound,
			expRes: httputils.JSONErrRes{
				Error: httputils.JSONErr{
					Message: "not found",
				},
			},
		},
		{
			label: "should pass with valid field",
			url:   url + "/1234",
			getAlbumFn: func(ctx context.Context, id string) (cl.GetAlbumRes, error) {
				return cl.GetAlbumRes{
					Album: &album,
				}, nil
			},
			expCode: http.StatusCreated,
			expRes: cl.GetAlbumRes{
				Album: &album,
			},
		},
	}
	for i := 0; i < len(table); i++ {
		ts := table[i]
		t.Run(ts.label, func(t *testing.T) {
			h := Handler{
				AlbumStore: &mock.AlbumStore{
					GetAlbumFn: ts.getAlbumFn,
				},
				Logger: tm.NopLogger,
			}

			h.Handler()
			wr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", ts.url, nil)
			h.router.ServeHTTP(wr, req)

			if wr.Code != ts.expCode {
				var res httputils.JSONErrRes
				err := jsonutils.Decode(wr.Body, &res)
				if err != nil {
					t.Fatalf("unexpected error returned from decoding response body: %s", err.Error())
				}

				t.Fatalf("unexpected response code returned: %s %s", cmp.Diff(ts.expCode, wr.Code), res.Error.Message)
			}

			if wr.Code != 200 {
				var res httputils.JSONErrRes
				err := jsonutils.Decode(wr.Body, &res)
				if err != nil {
					t.Fatalf("unexpected error returned from decoding response body: %s", err.Error())
				}

				if !cmp.Equal(res, ts.expRes) {
					t.Fatalf("unexpected response returned: %s", cmp.Diff(res, ts.expRes))
				}
			} else {
				var res cl.GetAlbumRes
				err := jsonutils.Decode(wr.Body, &res)
				if err != nil {
					t.Fatalf("unexpected error returned from decoding response body: %s", err.Error())
				}

				if !cmp.Equal(res, ts.expRes) {
					t.Fatalf("unexpected response returned: %s", cmp.Diff(res, ts.expRes))
				}
			}
		})
	}
}
