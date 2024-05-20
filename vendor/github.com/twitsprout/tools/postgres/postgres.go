package postgres

import (
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	// Blank import of postgres driver.
	_ "github.com/lib/pq"
)

// QueryWriter is an interface that is responsible for writing a byte or
// a slice of bytes to a query writer.
type QueryWriter interface {
	Write(p []byte) (int, error)
	WriteByte(c byte) error
}

// Options represents the required variables for starting a postgres instance.
type Options struct {
	DBName     string
	DisableSSL bool
	Host       string
	Password   string
	Port       int
	Username   string

	MaxConnLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

// New returns a new sql.DB instance backed by PostgreSQL, using the provided
// options.
func New(ops Options) (*sql.DB, error) {
	// Set pooling/reuse settings.
	if ops.MaxConnLifetime <= 0 {
		ops.MaxConnLifetime = 30 * time.Minute
	}
	if ops.MaxIdleConns <= 0 {
		ops.MaxIdleConns = 15
	}
	if ops.MaxOpenConns <= 0 {
		ops.MaxOpenConns = 20
	}
	return newDB(ops)
}

func newDB(ops Options) (*sql.DB, error) {
	connStr := connStrFromOptions(ops)

	// Open postgres connection.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set pooling/reuse settings.
	if ops.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(ops.MaxConnLifetime)
	}
	if ops.MaxIdleConns > 0 {
		db.SetMaxIdleConns(ops.MaxIdleConns)
	}
	if ops.MaxOpenConns > 0 {
		db.SetMaxOpenConns(ops.MaxOpenConns)
	}

	return db, nil
}

// NestedPlaceholders returns the string of nested placeholders with n number
// of entries of m values each.
//
// e.g. nestedPlaceholderText(2, 3, 1) = "($2,$3,$4),($5,$6,$7)"
func NestedPlaceholders(p QueryWriter, values, arguments, offset int) error {
	var err error
	for i := 0; i < values; i++ {
		if i > 0 {
			_ = p.WriteByte(',')
		}
		err = Placeholders(p, arguments, i*arguments+offset)
	}
	return err
}

// Placeholders returns the string of placeholders with n values and an
// offset of offset.
//
// e.g. placeholderText(3, 6) = "($7,$8,$9)"
func Placeholders(p QueryWriter, n, offset int) error {
	var err error
	var buf [64]byte
	_ = p.WriteByte('(')
	for i := 0; i < n; i++ {
		if i > 0 {
			_ = p.WriteByte(',')
		}
		_ = p.WriteByte('$')
		num := strconv.AppendInt(buf[:0], int64(i+offset+1), 10)
		_, err = p.Write(num)
	}
	_ = p.WriteByte(')')
	return err
}

func urlFromOptions(ops Options) string {
	// dbURL represents the connection URL for Postgres (to be formatted).
	const dbURL = "postgres://%s:%s@%s/%s?sslmode=%s"

	var sslmode string
	if ops.DisableSSL {
		sslmode = "disable"
	} else {
		sslmode = "require"
	}
	return fmt.Sprintf(dbURL, ops.Username, ops.Password, ops.Host, ops.DBName, sslmode)
}

// connStrFromOptions returns the libpq connection string given the provided
// options. Empty string values are not set in the returned string.
func connStrFromOptions(ops Options) string {
	var sslmode string
	if ops.DisableSSL {
		sslmode = "disable"
	} else {
		sslmode = "require"
	}

	var port string
	if ops.Port > 0 {
		port = strconv.Itoa(ops.Port)
	}

	var b strings.Builder
	writeConnStrParam(&b, "dbname", ops.DBName)
	writeConnStrParam(&b, "host", ops.Host)
	writeConnStrParam(&b, "password", ops.Password)
	writeConnStrParam(&b, "port", port)
	writeConnStrParam(&b, "sslmode", sslmode)
	writeConnStrParam(&b, "user", ops.Username)
	return b.String()
}

// writeConnStrParam writes the connection parameter to the provided
// StringWriter. If the provided value is empty, nothing is written.
func writeConnStrParam(w io.StringWriter, name, value string) { //nolint:interfacer
	if value == "" {
		return
	}
	_, _ = w.WriteString(name)
	_, _ = w.WriteString("='")
	_, _ = w.WriteString(strings.Replace(value, "'", `\'`, -1))
	_, _ = w.WriteString("' ")
}
