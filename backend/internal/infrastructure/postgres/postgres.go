package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the PostgreSQL connection pool
type DB struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL connection
func New(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute
	config.ConnConfig.ConnectTimeout = time.Second * 5

	// Set default query timeout
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.pool.Close()
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	if db.pool == nil {
		return errors.New("database connection not initialized")
	}
	return db.pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *DB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryRow executes a query that returns at most one row
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

// Query executes a query that returns multiple rows
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

// Exec executes a query without returning any rows
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}
