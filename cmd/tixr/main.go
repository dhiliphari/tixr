package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dhiliphari/tixr/internal/booking"
	"github.com/dhiliphari/tixr/internal/catalog"
	"github.com/jackc/pgx/v5/pgxpool"
)

var version = "dev"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config  config
	logger  *slog.Logger
	db      *pgxpool.Pool
	catalog *catalog.Store
	booking *booking.Store
}

func main() {
	cfg, displayVersion := loadCfg()

	if displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if cfg.db.dsn == "" {
		logger.Error("database DSN is required: set TIXR_DB_DSN or pass -db-dsn")
		os.Exit(1)
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		db:     db,
	}

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env, "version", version)

	err = srv.ListenAndServe()
	logger.Error("server stopped", "error", err)
	os.Exit(1)
}

func loadCfg() (config, bool) {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("TIXR_DB_DSN"), "PostgreSQL DSN")
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	return cfg, *displayVersion
}

func openDB(cfg config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
