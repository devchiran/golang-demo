package mock

import (
	"context"
	cl "golang-demo/pkg/catelog"
)

// AlbumStore defines an interface responsible for Album CRUD.
type AlbumStore struct {
	ListAlbumsFn  func(ctx context.Context) (cl.ListAlbumsRes, error)
	GetAlbumFn    func(ctx context.Context, id string) (cl.GetAlbumRes, error)
	CreateAlbumFn func(ctx context.Context, req cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error)
}

// ListAlbum proxies the request to the ListAlbum that's injected when
// the mock store is created.
func (s *AlbumStore) ListAlbums(ctx context.Context) (cl.ListAlbumsRes, error) {
	return s.ListAlbumsFn(ctx)
}

// CreateAlbum proxies the request to the CreateAlbum that's injected when
// the mock store is created.
func (s *AlbumStore) CreateAlbum(ctx context.Context, req cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error) {
	return s.CreateAlbumFn(ctx, req)
}

// GetAlbum proxies the request to the GetAlbum that's injected when
// the mock store is created.
func (s *AlbumStore) GetAlbum(ctx context.Context, req cl.GetAlbumReq) (cl.GetAlbumRes, error) {
	return s.GetAlbumFn(ctx, req)
}
