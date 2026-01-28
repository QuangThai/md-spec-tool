package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connected successfully")
	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "create_users",
			sql: `CREATE TABLE IF NOT EXISTS users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email VARCHAR(255) UNIQUE NOT NULL,
				password VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		},
		{
			name: "create_templates",
			sql: `CREATE TABLE IF NOT EXISTS templates (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				name VARCHAR(255) NOT NULL,
				content TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_templates_user_id ON templates(user_id);`,
		},
		{
			name: "create_specs",
			sql: `CREATE TABLE IF NOT EXISTS specs (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				template_id UUID REFERENCES templates(id),
				title VARCHAR(255) NOT NULL,
				content TEXT NOT NULL,
				version INTEGER DEFAULT 1,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_specs_user_id ON specs(user_id);
			CREATE INDEX IF NOT EXISTS idx_specs_template_id ON specs(template_id);`,
		},
		{
			name: "add_deleted_at_to_specs",
			sql:  `ALTER TABLE specs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;`,
		},
	}

	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
		log.Printf("✓ Migration %s applied", m.name)
	}

	return nil
}
