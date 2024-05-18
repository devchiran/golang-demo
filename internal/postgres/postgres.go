package postgres

import (
	"regexp"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/jmoiron/sqlx"
	"github.com/twitsprout/tools"
	"github.com/twitsprout/tools/postgres"
)

type Config postgres.Config

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// Postgres represents the type to interact with the PostgreSQL database.
type Postgres struct {
	sqldb *sqlx.DB
	db    *postgres.DB
}

type QueryValues struct {
	query string
	args  []interface{}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// New creates a new Postgres store.
func New(c Config, sc tools.StatsClient) (*Postgres, error) {
	db, err := postgres.NewDB(postgres.Config(c))
	if err != nil {
		return nil, err
	}
	sqldb := sqlx.NewDb(db.SQLDB(), "postgres")
	sqldb.MapperFunc(ToSnakeCase)
	if err != nil {
		return nil, err
	}
	return &Postgres{sqldb: sqldb, db: db}, nil
}
