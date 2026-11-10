package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
	"github.com/dhiliphari/tixr/internal/catalog"
	"github.com/dhiliphari/tixr/internal/database"
	"github.com/dhiliphari/tixr/internal/grpcserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var version = "dev"

type config struct {
	port            int
	env             string
	shutdownTimeout time.Duration
	db              struct {
		dsn string
	}
}

type application struct {
	catalogv1.UnimplementedCatalogServiceServer

	config config
	logger *slog.Logger
	store  *catalog.Store
}

func main() {
	cfg, displayVersion := loadCfg()

	if displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "catalog")

	if err := run(cfg, logger); err != nil {
		logger.Error("catalog exited with an error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config, logger *slog.Logger) error {
	if cfg.db.dsn == "" {
		return errors.New("database DSN is required: set TIXR_DB_DSN or pass -db-dsn")
	}

	openCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(openCtx, cfg.db.dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		store:  catalog.NewStore(db),
	}

	srv := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(srv, app)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	reflection.Register(srv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	context.AfterFunc(ctx, stop)

	logger.Info("catalog starting", "env", cfg.env, "version", version)
	return grpcserver.Run(ctx, logger, srv, hs, fmt.Sprintf(":%d", cfg.port), cfg.shutdownTimeout)
}

func loadCfg() (config, bool) {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("TIXR_DB_DSN"), "PostgreSQL DSN")
	flag.DurationVar(&cfg.shutdownTimeout, "shutdown-timeout", 20*time.Second, "Maximum time to wait for in-flight work during shutdown")
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	return cfg, *displayVersion
}
