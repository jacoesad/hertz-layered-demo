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
			subtask_type TEXT NOT NULL DEFAULT 'manual',
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
	if err := ensureSubtaskTypeColumn(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensureSubtaskTypeColumn(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(subtasks)`)
	if err != nil {
		return fmt.Errorf("inspect subtasks schema: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("scan subtasks schema: %w", err)
		}
		if name == "subtask_type" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate subtasks schema: %w", err)
	}

	if _, err := db.ExecContext(ctx, `ALTER TABLE subtasks ADD COLUMN subtask_type TEXT NOT NULL DEFAULT 'manual'`); err != nil {
		return fmt.Errorf("add subtasks.subtask_type: %w", err)
	}
	return nil
}

func seed(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`INSERT OR IGNORE INTO tasks (id, tenant_id, title, status, owner) VALUES (1001, 'tenant-a', 'Prepare onboarding', 'running', 'Alice')`,
		`INSERT OR IGNORE INTO tasks (id, tenant_id, title, status, owner) VALUES (1001, 'tenant-b', 'Sync external ticket', 'done', 'Bob')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, subtask_type, title, status, assignee) VALUES (5001, 'tenant-a', 1001, 'manual', 'Collect documents', 'todo', 'Cindy')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, subtask_type, title, status, assignee) VALUES (5002, 'tenant-a', 1001, 'review', 'Review checklist', 'running', 'Evan')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, subtask_type, title, status, assignee) VALUES (5001, 'tenant-b', 1001, 'webhook', 'Confirm webhook', 'done', 'David')`,
		`UPDATE subtasks SET subtask_type = 'manual' WHERE tenant_id = 'tenant-a' AND id = 5001`,
		`UPDATE subtasks SET subtask_type = 'review' WHERE tenant_id = 'tenant-a' AND id = 5002`,
		`UPDATE subtasks SET subtask_type = 'webhook' WHERE tenant_id = 'tenant-b' AND id = 5001`,
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
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, subtask_type, title, status, assignee) VALUES (9001, 'tenant-a', 1001, 'metric', 'StarRocks aggregated subtask count', 'ready', 'analytics')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, subtask_type, title, status, assignee) VALUES (9002, 'tenant-a', 1001, 'metric', 'StarRocks latency snapshot', 'ready', 'analytics')`,
		`INSERT OR IGNORE INTO subtasks (id, tenant_id, task_id, subtask_type, title, status, assignee) VALUES (9001, 'tenant-b', 1001, 'sync', 'StarRocks external sync summary', 'ready', 'analytics')`,
		`UPDATE subtasks SET subtask_type = 'metric' WHERE tenant_id = 'tenant-a' AND id IN (9001, 9002)`,
		`UPDATE subtasks SET subtask_type = 'sync' WHERE tenant_id = 'tenant-b' AND id = 9001`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("seed starrocks database: %w", err)
		}
	}
	return nil
}
