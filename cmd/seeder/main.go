package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/matthewsaunders/link-shortener-api/internal/data"
	"github.com/rs/zerolog"
)

type config struct {
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *zerolog.Logger
	models data.Models
}

func main() {
	fmt.Println("-- main")

	var cfg config

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://shrtnr:password@localhost/shrtnr?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max open idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().Logger()

	/*
	 * Setup database connection
	 */
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal().Err(err).Msg("")
		}
	}()

	// Seed database
	seeder := &seeder{
		filename: "data.json",
		logger:   &logger,
		models:   data.NewModels(db),
	}

	fmt.Println("-- main.seedDB")
	err = seeder.seedDB()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to seed DB")
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
