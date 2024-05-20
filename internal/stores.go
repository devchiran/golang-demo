package internal

import (
	"context"
	cl "golang-demo/pkg/catelog"
)

type AlbumStore interface {
	ListAlbums(context.Context) (cl.ListAlbumsRes, error)
	GetAlbum(ctx context.Context, id string) (cl.GetAlbumRes, error)
}
