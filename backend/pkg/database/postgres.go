package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/SidahmedSeg/document-manager/backend/pkg/config"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"go.uber.org/zap"
)

// DB wraps sql.DB with additional methods
type DB struct {
	*sql.DB
	logger *zap.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.DatabaseConfig, logger *zap.Logger) (*DB, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to open database", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to ping database", err)
	}

	if logger != nil {
		logger.Info("database connection established",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.Name),
		)
	}

	return &DB{
		DB:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.logger != nil {
		db.logger.Info("closing database connection")
	}
	return db.DB.Close()
}

// HealthCheck performs a health check on the database
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(errors.ErrCodeDatabase, "database health check failed", err)
	}

	// Check if we can execute a simple query
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return errors.Wrap(errors.ErrCodeDatabase, "database query check failed", err)
	}

	return nil
}

// Stats returns database statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// TxFunc is a function that runs within a transaction
type TxFunc func(*sql.Tx) error

// WithTransaction executes a function within a database transaction
func (db *DB) WithTransaction(ctx context.Context, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(errors.ErrCodeDatabase, "failed to begin transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			if db.logger != nil {
				db.logger.Error("failed to rollback transaction",
					zap.Error(rbErr),
					zap.Error(err),
				)
			}
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(errors.ErrCodeDatabase, "failed to commit transaction", err)
	}

	return nil
}

// ExecContext executes a query with context and error wrapping
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		if db.logger != nil {
			db.logger.Error("query execution failed",
				zap.String("query", query),
				zap.Error(err),
			)
		}
		return nil, errors.Wrap(errors.ErrCodeDatabase, "query execution failed", err)
	}
	return result, nil
}

// QueryContext executes a query with context and error wrapping
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		if db.logger != nil {
			db.logger.Error("query failed",
				zap.String("query", query),
				zap.Error(err),
			)
		}
		return nil, errors.Wrap(errors.ErrCodeDatabase, "query failed", err)
	}
	return rows, nil
}

// QueryRowContext executes a query that returns a single row
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// SetTenantContext sets the tenant ID in the PostgreSQL session
// This can be used with Row Level Security (RLS) policies
func SetTenantContext(ctx context.Context, tx *sql.Tx, tenantID string) error {
	query := fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID)
	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(errors.ErrCodeDatabase, "failed to set tenant context", err)
	}
	return nil
}

// GetTenantFromContext retrieves tenant ID from PostgreSQL session
func GetTenantFromContext(ctx context.Context, db *sql.DB) (string, error) {
	var tenantID sql.NullString
	err := db.QueryRowContext(ctx, "SELECT current_setting('app.tenant_id', true)").Scan(&tenantID)
	if err != nil {
		return "", errors.Wrap(errors.ErrCodeDatabase, "failed to get tenant context", err)
	}

	if !tenantID.Valid {
		return "", nil
	}

	return tenantID.String, nil
}

// Helper functions for common operations

// Exists checks if a record exists
func Exists(ctx context.Context, db *sql.DB, query string, args ...interface{}) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(errors.ErrCodeDatabase, "exists check failed", err)
	}
	return exists, nil
}

// Count returns the count of records
func Count(ctx context.Context, db *sql.DB, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeDatabase, "count query failed", err)
	}
	return count, nil
}

// ScanOne scans a single row into dest
func ScanOne(rows *sql.Rows, dest ...interface{}) error {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return errors.Wrap(errors.ErrCodeDatabase, "scan failed", err)
		}
		return errors.ErrNotFound
	}

	if err := rows.Scan(dest...); err != nil {
		return errors.Wrap(errors.ErrCodeDatabase, "scan failed", err)
	}

	return nil
}

// BuildWhereClause builds a WHERE clause from a map of conditions
func BuildWhereClause(conditions map[string]interface{}) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var whereParts []string
	var args []interface{}
	argIndex := 1

	for key, value := range conditions {
		whereParts = append(whereParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	whereClause := "WHERE " + whereParts[0]
	for i := 1; i < len(whereParts); i++ {
		whereClause += " AND " + whereParts[i]
	}

	return whereClause, args
}
