package internal

import (
	"context"
	cl "golang-demo/pkg/catelog"
)

type AlbumStore interface {
	ListAlbums(ctx context.Context) (cl.ListAlbumsRes, error)
	GetAlbum(ctx context.Context, id string) (cl.GetAlbumRes, error)
	CreateAlbum(ctx context.Context, req cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error)
}
