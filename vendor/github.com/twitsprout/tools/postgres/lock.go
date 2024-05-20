package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/twitsprout/tools/distlock"
)

// Lock attempts to obtain the lock for the provided lockID. If the lock is
// already held by another service, ErrLockNotHeld should be returned.
func Lock(ctx context.Context, db *sql.DB, instanceID, lockID string, ttlSeconds int) error {
	const query = `
		UPDATE locks
		SET service_id = $1, last_op = 'lock', valid_until = now() + $2
		WHERE name = $3 AND (last_op = 'unlock' OR valid_until < now())
		RETURNING valid_until`

	var validUntil sql.NullString
	ttlStr := fmt.Sprintf("%d seconds", ttlSeconds)
	err := db.QueryRowContext(ctx, query, instanceID, ttlStr, lockID).Scan(&validUntil)
	if err == sql.ErrNoRows {
		return distlock.ErrLockNotHeld
	}
	return err
}

// Unlock releases the lock for provided instance/lock ID. If the lock is not
// held by the current service, ErrLockNotHeld should be returned.
func Unlock(ctx context.Context, db *sql.DB, instanceID, lockID string) error {
	const query = `
		UPDATE locks
		SET valid_until = NULL, service_id = '', last_op = 'unlock'
		WHERE name = $1 AND service_id = $2 AND last_op IN ('lock', 'extend') AND valid_until > now()
		RETURNING valid_until`

	var validUntil sql.NullString
	err := db.QueryRowContext(ctx, query, lockID, instanceID).Scan(&validUntil)
	if err == sql.ErrNoRows {
		return distlock.ErrLockNotHeld
	}
	return err
}

// Extend extends the TTL of the provided lock. If the lock isn't held by the
// instanceID, ErrLockNotHeld should be returned.
func Extend(ctx context.Context, db *sql.DB, instanceID, lockID string, ttlSeconds int) error {
	const query = `
		UPDATE locks
		SET last_op = 'extend', valid_until = now() + $1
		WHERE name = $2 AND service_id = $3 AND last_op IN ('lock', 'extend')
		RETURNING valid_until`

	var validUntil sql.NullString
	ttlStr := fmt.Sprintf("%d seconds", ttlSeconds)
	err := db.QueryRowContext(ctx, query, ttlStr, lockID, instanceID).Scan(&validUntil)
	if err == sql.ErrNoRows {
		return distlock.ErrLockNotHeld
	}
	return err
}
