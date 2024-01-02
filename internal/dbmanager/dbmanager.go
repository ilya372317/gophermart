package dbmanager

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"

	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseDSN string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed open database connection: %w", err)
	}
	runMigrations(db.DB)

	return db, nil
}

func runMigrations(db *sql.DB) {
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		logger.Log.Fatalf("failed init postgres driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations",
		"metrics", driver)
	if err != nil {
		logger.Log.Fatalf("failed get migration instance: %v", err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			logger.Log.Fatalf("failed run migrations: %v", err)
		}
	}
}

func MakeTestConnection() (*sqlx.DB, *dockertest.Pool, *dockertest.Resource, error) {
	var db *sqlx.DB

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	resource, err := pool.Run("postgres", "15", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=gopher_test"})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	port := resource.GetPort("5432/tcp")
	connectionString := fmt.Sprintf(
		"host=localhost port=%s user=postgres password=secret dbname=gopher_test sslmode=disable",
		port,
	)

	if err = pool.Retry(func() error {
		var err error
		db, err = sqlx.Open("pgx", connectionString)
		if err != nil {
			return fmt.Errorf("failed open test connection: %w", err)
		}
		pingErr := db.Ping()
		if pingErr != nil {
			return fmt.Errorf("failed ping test db: %w", pingErr)
		}
		return nil
	}); err != nil {
		return nil, nil, nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	return db, pool, resource, nil
}

func CloseTestConnection(db *sqlx.DB, pool *dockertest.Pool, resource *dockertest.Resource) error {
	_ = db.Close()
	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("failed purge docker resource: %w", err)
	}

	return nil
}
