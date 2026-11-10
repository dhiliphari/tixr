package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
	"github.com/dhiliphari/tixr/internal/booking"
	"github.com/dhiliphari/tixr/internal/database"
	"github.com/dhiliphari/tixr/internal/grpcserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var version = "dev"

type config struct {
	port            int
	env             string
	catalogAddr     string
	shutdownTimeout time.Duration
	mailDelay       time.Duration
	db              struct {
		dsn string
	}
}

type application struct {
	bookingv1.UnimplementedBookingServiceServer

	config  config
	logger  *slog.Logger
	store   *booking.Store
	catalog catalogv1.CatalogServiceClient
	mailer  *mailer
	now     func() time.Time
	wg      sync.WaitGroup
}

func main() {
	cfg, displayVersion := loadCfg()

	if displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "booking")

	if err := run(cfg, logger); err != nil {
		logger.Error("booking exited with an error", "error", err)
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

	catalogConn, err := grpc.NewClient(cfg.catalogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create catalog client: %w", err)
	}
	defer catalogConn.Close()

	app := &application{
		config:  cfg,
		logger:  logger,
		store:   booking.NewStore(db),
		catalog: catalogv1.NewCatalogServiceClient(catalogConn),
		mailer:  &mailer{logger: logger, delay: cfg.mailDelay},
		now:     time.Now,
	}

	srv := grpc.NewServer()
	bookingv1.RegisterBookingServiceServer(srv, app)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	reflection.Register(srv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	context.AfterFunc(ctx, stop)

	logger.Info("booking starting", "env", cfg.env, "version", version)
	runErr := grpcserver.Run(ctx, logger, srv, hs, fmt.Sprintf(":%d", cfg.port), cfg.shutdownTimeout)

	// RPC handlers start background tasks, so only wait for those once every
	// handler has finished
	waitCtx, waitCancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
	defer waitCancel()

	logger.Info("waiting for background tasks")
	waitErr := app.waitBackground(waitCtx)

	return errors.Join(runErr, waitErr)
}

func loadCfg() (config, bool) {
	var cfg config

	flag.IntVar(&cfg.port, "port", 50052, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("TIXR_DB_DSN"), "PostgreSQL DSN")
	flag.StringVar(&cfg.catalogAddr, "catalog-addr", "localhost:50051", "Catalog service address")
	flag.DurationVar(&cfg.mailDelay, "mail-delay", 2*time.Second, "Simulated latency of sending an email")
	flag.DurationVar(&cfg.shutdownTimeout, "shutdown-timeout", 20*time.Second, "Maximum time to wait for in-flight work during shutdown")
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	return cfg, *displayVersion
}
