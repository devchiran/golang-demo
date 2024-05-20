package postgres

import (
	"context"
	cl "golang-demo/pkg/catelog"

	sq "github.com/Masterminds/squirrel"

	"github.com/pkg/errors"
)

const tableAlbums = "albums"

const (
	albumsColumnID        = `"id"`
	albumsColumnTitle     = `"title"`
	albumsColumnCreatedAt = `"created_at"`
	albumsColumnUpdatedAt = `"updated_at"`
)

var albumsColumns = []string{
	albumsColumnID,
	albumsColumnTitle,
	albumsColumnCreatedAt,
	albumsColumnUpdatedAt,
}

func (p *Postgres) ListAlbums(ctx context.Context) (cl.ListAlbumsRes, error) {

	var res cl.ListAlbumsRes

	var r []cl.Album
	qv, err := buildListAlbumsQuery()
	if err != nil {
		return res, errors.Wrap(err, "build list albums query")
	}
	err = p.sqldb.SelectContext(ctx, &r, qv.query, qv.args...)
	if err != nil {
		return res, errors.Wrap(err, "execute list albums query")
	}

	// If not rows are found, return a 404.
	if len(r) == 0 {
		return res, cl.ErrNotFound
	}

	res = cl.ListAlbumsRes{
		Albums: r,
	}
	return res, nil

}

func buildListAlbumsQuery() (QueryValues, error) {
	q, args, err := psql.
		Select(tableColumns(tableAlbums, albumsColumns)...).
		From(tableAlbums).
		OrderBy(albumsColumnCreatedAt + " DESC").
		ToSql()

	return QueryValues{q, args}, errors.Wrap(err, "list albums build query into SQL string")
}

func (p *Postgres) GetAlbum(ctx context.Context, id string) (cl.ListAlbumsRes, error) {

	var res cl.ListAlbumsRes

	var r []cl.Album
	qv, err := buildGetAlbumQuery(id)
	if err != nil {
		return res, errors.Wrap(err, "build get album query")
	}
	err = p.sqldb.SelectContext(ctx, &r, qv.query, qv.args...)
	if err != nil {
		return res, errors.Wrap(err, "execute get album query")
	}

	// If not rows are found, return a 404.
	if len(r) == 0 {
		return res, cl.ErrNotFound
	}

	res = cl.GetAlbumRes{
		Album: r,
	}
	return res, nil

}

func buildGetAlbumQuery(id string) (QueryValues, error) {
	q, args, err := psql.
		Select(tableColumns(tableAlbums, albumsColumns)...).
		From(tableAlbums).
		Where(sq.Eq{"id": id}).
		ToSql()

	return QueryValues{q, args}, errors.Wrap(err, "get album build query into SQL string")
}
