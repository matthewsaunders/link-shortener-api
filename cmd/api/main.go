package main

import (
	"flag"
	"os"
	"sync"

	"github.com/matthewsaunders/link-shortener-api/internal/vcs"
	"github.com/rs/zerolog"
)

var version = vcs.Version()

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *zerolog.Logger
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	/*
	 *	Parse command line options
	 */
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max open idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().Logger()

	app := &application{
		config: cfg,
		logger: &logger,
	}

	if err := app.serve(); err != nil {
		logger.Fatal().Err(err)
	}
}
