package postgres

import (
	"context"
	"database/sql"
	cl "golang-demo/pkg/catelog"
	"strings"

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

func (p *Postgres) GetAlbum(ctx context.Context, id string) (cl.GetAlbumRes, error) {

	var res cl.GetAlbumRes

	var r cl.Album
	qv, err := buildGetAlbumQuery(id)
	if err == sql.ErrNoRows {
		return res, cl.ErrNotFound
	}
	if err != nil {
		return res, errors.Wrap(err, "build get album query")
	}
	err = p.sqldb.SelectContext(ctx, &r, qv.query, qv.args...)
	if err != nil {
		return res, errors.Wrap(err, "execute get album query")
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

func (p *Postgres) CreateAlbum(ctx context.Context, req cl.CreateAlbumRequest) (cl.CreateAlbumResponse, error) {
	var res cl.CreateAlbumResponse

	var r cl.Album
	qv, err := buildCreateAlbumQuery(req)
	if err != nil {
		return res, errors.Wrap(err, "build Album insert query")
	}
	err = p.sqldb.GetContext(ctx, &r, qv.query, qv.args...)
	if err != nil {
		return res, errors.Wrap(err, "execute Album insert query")
	}

	res = cl.CreateAlbumResponse{
		Album: &r,
	}
	return res, nil
}

func buildCreateAlbumQuery(req cl.CreateAlbumRequest) (QueryValues, error) {
	q, args, err := psql.Insert(tableAlbums).
		Columns(albumsColumnID, albumsColumnTitle, albumsColumnCreatedAt, albumsColumnUpdatedAt).
		Values(req.AlbumID, req.Title, "now()", "NULL").
		Suffix("RETURNING " + strings.Join(albumsColumns, " , ")).
		ToSql()

	return QueryValues{q, args}, errors.Wrap(err, "create album build query into SQL string")
}
