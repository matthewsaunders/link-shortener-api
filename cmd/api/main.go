package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/matthewsaunders/link-shortener-api/internal/data"
	"github.com/matthewsaunders/link-shortener-api/internal/vcs"
	"github.com/rs/zerolog"
)

var version = vcs.Version()

type config struct {
	port      int
	env       string
	migrateDB bool
	db        struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *zerolog.Logger
	models data.Models
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	/*
	 *	Parse command line options
	 */
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production")
	flag.BoolVar(&cfg.migrateDB, "migrate-db", false, "Run DB migrations")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://shrtnr:password@localhost/shrtnr?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max open idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().Logger()

	/*
	 * Setup database connection
	 */
	logger.Info().Msg("Opening DB connection")
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal().Err(err).Msg("")
		}
	}()

	/*
	 * Migrate DB
	 */
	logger.Info().Msg("Migrating DB")
	if cfg.migrateDB {
		err = migrateDB(db)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to migrate DB")
		}
	}

	/*
	 * Start application server
	 */
	logger.Info().Msg("Starting application")
	app := &application{
		config: cfg,
		logger: &logger,
		models: data.NewModels(db),
	}

	if err := app.serve(); err != nil {
		logger.Fatal().Err(err)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	// Note that passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connection in the pool. Again,
	// passing a value less than or equal to 0 will mean there is no limit
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout duration string to a
	// time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database,
	// passing in the context we created above as a parameter.
	// If connection couldn't be established successfully within the 5-second deadline,
	// then this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}

func migrateDB(db *sql.DB) error {
	migrationDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://migrations", "shrtnr", migrationDriver)
	if err != nil {
		return err
	}

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
