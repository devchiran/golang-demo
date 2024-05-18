package internal

import (
	"context"
	cl "golang-demo/pkg/catelog"
)

type AlbumStore interface {
	ListAlbums(context.Context) (cl.ListAlbumRes, error)
}
