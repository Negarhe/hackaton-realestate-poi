package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       //  'file' source driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFileSystem embed.FS

func ConnectToDatabase(
	serverContext context.Context,
	databaseUsername string,
	databasePassword string,
	databaseHost string,
	databaseName string,
	databasePort int,
	databaseSSLMode string,
	databaseMaxConns int32,
	databaseMinConns int32,
	databaseMaxConnLifetimeJitterMinutes int,
	databaseMaxConnLifetimeMinutes int,
	databaseMaxConnIdleTimeMinutes int,
) (*pgxpool.Pool, error) {

	//check if config is provided
	if databaseHost == "" || databaseName == "" || databaseUsername == "" || databasePassword == "" || databasePort == 0 {
		return nil, fmt.Errorf("incomplete database config: userName: %s password: %s, host: %s, name: %s, port: %d",
			databaseUsername, databasePassword, databaseHost, databaseName, databasePort)
	}

	//create database url
	// example URL: postgres://username:password@localhost:5432/database_name
	pgURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(databaseUsername, databasePassword),
		Host:   fmt.Sprintf("%s:%d", databaseHost, databasePort),
		Path:   "/" + databaseName,
	}
	query := pgURL.Query()
	query.Add("sslmode", databaseSSLMode)
	// fmt.Printf(query.Encode() + "\n")
	pgURL.RawQuery = query.Encode()
	pgxPoolConfig, err := pgxpool.ParseConfig(pgURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config URL: %w", err)
	}
	// Configure pgxpool Parameters
	pgxPoolConfig.MaxConns = databaseMaxConns
	pgxPoolConfig.MinConns = databaseMinConns
	pgxPoolConfig.MaxConnLifetime = time.Duration(databaseMaxConnLifetimeMinutes) * time.Minute
	pgxPoolConfig.MaxConnIdleTime = time.Duration(databaseMaxConnIdleTimeMinutes) * time.Minute
	pgxPoolConfig.MaxConnLifetimeJitter = time.Duration(databaseMaxConnLifetimeJitterMinutes) * time.Minute

	// Establish Connection Pool
	pool, err := pgxpool.NewWithConfig(serverContext, pgxPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create PGX connection pool: %w", err)
	}

	// Ping the Database to Verify Connection
	if err := pool.Ping(serverContext); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to reach database: %w", err)
	}
	// m, err := migrate.New("file:///home/divar/Realestate-POI/pkg/database/migrations", pgURL.String())
	// if err != nil {
	// 	pool.Close()
	// 	return nil, fmt.Errorf("migration initialization failed: %w", err)
	// }
	driver, err := iofs.New(migrationFileSystem, "migrations")
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to load migrations: %w", err)

	}
	m, err := migrate.NewWithSourceInstance("iofs", driver, pgURL.String())
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("migration initialization failed: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		pool.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	log.Println("Connected to the database successfully.")
	return pool, nil

}
