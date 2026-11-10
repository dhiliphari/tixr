package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/dhiliphari/tixr/internal/booking"
	"github.com/dhiliphari/tixr/internal/catalog"
	"github.com/dhiliphari/tixr/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

var version = "dev"

type config struct {
	port            int
	env             string
	mailDelay       time.Duration
	shutdownTimeout time.Duration
	db              struct {
		dsn string
	}
}

type application struct {
	config  config
	logger  *slog.Logger
	db      *pgxpool.Pool
	catalog *catalog.Store
	booking *booking.Store
	mailer  *mailer
	wg      sync.WaitGroup
}

func main() {
	cfg, displayVersion := loadCfg()

	if displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(cfg, logger); err != nil {
		logger.Error("tixr exited with an error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config, logger *slog.Logger) error {
	if cfg.db.dsn == "" {
		return errors.New("database DSN is required: set TIXR_DB_DSN or pass -db-dsn")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.db.dsn)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config:  cfg,
		logger:  logger,
		db:      db,
		catalog: catalog.NewStore(db),
		booking: booking.NewStore(db),
		mailer:  &mailer{logger: logger, delay: cfg.mailDelay},
	}

	return app.serve()
}

func loadCfg() (config, bool) {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("TIXR_DB_DSN"), "PostgreSQL DSN")
	flag.DurationVar(&cfg.mailDelay, "mail-delay", 2*time.Second, "Simulated latency of sending an email")
	flag.DurationVar(&cfg.shutdownTimeout, "shutdown-timeout", 20*time.Second, "Maximum time to wait for in-flight work during shutdown")
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	return cfg, *displayVersion
}
