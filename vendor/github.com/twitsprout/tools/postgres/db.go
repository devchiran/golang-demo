package postgres

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Scanner represents the interface for scanning the result of a returned row
// into acceptable Go type(s). This interface is used in the QueryRowPrepared
// method of a Conn.
type Scanner interface {
	Scan(dest ...interface{}) error
}

// Conn is the interface for a connection to postgres exposed by the DB's Do
// method. It includes most methods on a *sql.DB instance, as well as three new
// methods (ExecPrepared, QueryPrepared, and QueryRowPrepared) that utilize a
// cache of prepared statements, increasing performance ~2x in most cases.
type Conn interface {
	Begin() (*sql.Tx, error)
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	Exec(string, ...interface{}) (sql.Result, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	ExecPrepared(context.Context, string, ...interface{}) (sql.Result, error)
	Ping() error
	PingContext(context.Context) error
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryPrepared(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	QueryRowPrepared(context.Context, string, ...interface{}) Scanner
}

// DB is a wrapper around a *sql.DB, where users should call the Do method to
// execute queries in a safe manner. When finished with the DB, the Close method
// must be called to free all resources. If needed, the underlying *sql.DB
// instance can be accessed with the SQLDB method.
type DB struct {
	clock      Clock
	conn       *dbConn
	onComplete func(ctx context.Context, label string, start time.Time, err error) error
	semaphore  Semaphore
	timeout    time.Duration
}

// NewDB returns an initialized DB instance, using the provided Config, and any
// optional Options provided.
func NewDB(c Config, ops ...Option) (*DB, error) {
	o := defaultOptions()
	for _, op := range ops {
		op(&o)
	}

	db, err := newDB(Options{
		DBName:          c.Name,
		DisableSSL:      c.DisableSSL,
		Host:            c.Host,
		Password:        c.Password,
		Port:            c.Port,
		Username:        c.Username,
		MaxConnLifetime: o.maxConnLifetime,
		MaxIdleConns:    o.maxIdleConns,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		clock: o.clock,
		conn: &dbConn{
			DB: db,
			sf: &singleflight.Group{},
		},
		onComplete: o.onComplete,
		semaphore:  o.semaphore,
		timeout:    o.timeout,
	}, nil
}

// SQLDB returns the underlying *sql.DB instance used. This should only be used
// in cases where the caller MUST access methods not available on the Conn
// provided by calling the Do method.
func (db *DB) SQLDB() *sql.DB {
	return db.conn.DB
}

// Close closes all cached prepared statements, and then closes the underlying
// *sql.DB instance. Close must be called whenever the the DB object is no
// longer used to free all resources.
func (db *DB) Close() error {
	db.conn.closeAll()
	return db.conn.DB.Close()
}

// Do is the method that should be used to execute a query on the underlying
// database. It accepts a parent context, a label for the operation, and a
// function that will be invoked with a context and Conn, returning any error
// that is encountered. The provided Conn should be used to execute queries, and
// must not be retained outside of the function scope.
func (db *DB) Do(ctx context.Context, label string, fn func(context.Context, Conn) error) (err error) {
	if db.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, db.timeout)
		defer cancel()
	}

	if db.semaphore != nil {
		if err = db.semaphore.Acquire(ctx); err != nil {
			return
		}
		defer db.semaphore.Release()
	}

	if db.onComplete != nil {
		start := db.now()
		defer func() {
			err = db.onComplete(ctx, label, start, err)
		}()
	}

	err = fn(ctx, db.conn)
	return
}

func (db *DB) now() time.Time {
	if db.clock == nil {
		return time.Now()
	}
	return db.clock.Now()
}

// dbConn represents the underlying type provided to the caller of the DB's Do
// method. It satisfies the Conn interface defined in this package.
// dbConn keeps a cache of prepared statements for increased performance, only
// preparing a statement for a query once.
type dbConn struct {
	*sql.DB

	sf *singleflight.Group

	mu    sync.RWMutex
	stmts map[string]*sql.Stmt // TODO(fowler): Consider using sync.Map here?
}

func (c *dbConn) ExecPrepared(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := c.getStmt(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

func (c *dbConn) QueryPrepared(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := c.getStmt(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

func (c *dbConn) QueryRowPrepared(ctx context.Context, query string, args ...interface{}) Scanner {
	stmt, err := c.getStmt(ctx, query)
	if err != nil {
		return &errScanner{err: err}
	}
	return stmt.QueryRowContext(ctx, args...)
}

// closeAll closes and removes all open prepared statements in the dbConn's
// cache.
func (c *dbConn) closeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, stmt := range c.stmts {
		_ = stmt.Close()
	}
	c.stmts = nil
}

// getStmt attempts to retrieve a cached prepared statement, falling back to
// creating one itself. Only one prepared statement per query should exist at
// any point in time.
func (c *dbConn) getStmt(ctx context.Context, query string) (*sql.Stmt, error) {
	// Fast path. Stmt already exists.
	c.mu.RLock()
	stmt, ok := c.stmts[query]
	c.mu.RUnlock()
	if ok && stmt != nil {
		return stmt, nil
	}

	// Use singleflight to prepare the statement only once.
	chRes := c.sf.DoChan(query, func() (interface{}, error) {
		// Check to see if stmt now exists before preparing.
		c.mu.RLock()
		stmt, ok := c.stmts[query]
		c.mu.RUnlock()
		if ok && stmt != nil {
			return stmt, nil
		}

		stmt, err := c.DB.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}

		// Save stmt in map before returning.
		c.mu.Lock()
		if c.stmts == nil {
			c.stmts = make(map[string]*sql.Stmt)
		}
		c.stmts[query] = stmt
		c.mu.Unlock()

		return stmt, nil
	})

	// Wait for the result of the singleflight func above.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-chRes:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*sql.Stmt), nil
	}
}

type errScanner struct {
	err error
}

func (s *errScanner) Scan(_ ...interface{}) error {
	return s.err
}
