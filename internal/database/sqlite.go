package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"hz-server/internal/config"
)

type DataSources struct {
	DB          *sql.DB
	StarRocksDB *sql.DB
}

func Init(ctx context.Context, cfg *config.Config) (*DataSources, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	db, err := Open(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	starRocksDB, err := OpenStarRocks(ctx, cfg.StarRocks)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &DataSources{
		DB:          db,
		StarRocksDB: starRocksDB,
	}, nil
}

func (d *DataSources) Close() error {
	if d == nil {
		return nil
	}
	if d.StarRocksDB != nil {
		_ = d.StarRocksDB.Close()
	}
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

func Open(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	return open(ctx, cfg, seed)
}

func OpenStarRocks(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	return open(ctx, cfg, seedStarRocks)
}

func open(ctx context.Context, cfg config.DatabaseConfig, seedFunc func(context.Context, *sql.DB) error) (*sql.DB, error) {
	if cfg.Driver != "sqlite3" {
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}

	if err := ensureDir(cfg.DSN); err != nil {
		return nil, err
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	if err := seedFunc(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ensureDir(dsn string) error {
	if dsn == "" || dsn == ":memory:" {
		return nil
	}
	dir := filepath.Dir(dsn)
	if dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create database dir: %w", err)
	}
	return nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER NOT NULL,
			tenant_id TEXT NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			owner TEXT NOT NULL,
			PRIMARY KEY (tenant_id, id)
		)`,
		`CREATE TABLE IF NOT EXISTS subtasks (
			id INTEGER NOT NULL,
			tenant_id TEXT NOT NULL,
			task_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			assignee TEXT NOT NULL,
			PRIMARY KEY (tenant_id, id)
		)`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}
	return nil
}

func seed(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`INSERT OR IGNORE INTO tasks (id, tenant_id, title, status, owner) VALUES (1001, 'tenant-a', 'Prepare onboarding', 'running', 'Alice')`,
		`INSERT OR IGNORE INTO tasks (id, tenant_id, title, status, owner) VALUES (1001, 'tenant-b', 'Sync external ticket', 'done', 'Bob')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, title, status, assignee) VALUES (5001, 'tenant-a', 1001, 'Collect documents', 'todo', 'Cindy')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, title, status, assignee) VALUES (5002, 'tenant-a', 1001, 'Review checklist', 'running', 'Evan')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, title, status, assignee) VALUES (5001, 'tenant-b', 1001, 'Confirm webhook', 'done', 'David')`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("seed database: %w", err)
		}
	}
	return nil
}

func seedStarRocks(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, title, status, assignee) VALUES (9001, 'tenant-a', 1001, 'StarRocks aggregated subtask count', 'ready', 'analytics')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, title, status, assignee) VALUES (9002, 'tenant-a', 1001, 'StarRocks latency snapshot', 'ready', 'analytics')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, title, status, assignee) VALUES (9001, 'tenant-b', 1001, 'StarRocks external sync summary', 'ready', 'analytics')`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("seed starrocks database: %w", err)
		}
	}
	return nil
}
