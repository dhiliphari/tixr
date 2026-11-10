package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	bookingv1 "github.com/dhiliphari/tixr/gen/booking/v1"
	catalogv1 "github.com/dhiliphari/tixr/gen/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var version = "dev"

type config struct {
	port            int
	env             string
	catalogAddr     string
	bookingAddr     string
	shutdownTimeout time.Duration
}

type application struct {
	config  config
	logger  *slog.Logger
	catalog catalogv1.CatalogServiceClient
	booking bookingv1.BookingServiceClient
}

func main() {
	cfg, displayVersion := loadCfg()

	if displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "gateway")

	if err := run(cfg, logger); err != nil {
		logger.Error("tixr exited with an error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config, logger *slog.Logger) error {
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())

	catalogConn, err := grpc.NewClient(cfg.catalogAddr, creds)
	if err != nil {
		return fmt.Errorf("create catalog client: %w", err)
	}
	defer catalogConn.Close()

	bookingConn, err := grpc.NewClient(cfg.bookingAddr, creds)
	if err != nil {
		return fmt.Errorf("create booking client: %w", err)
	}
	defer bookingConn.Close()

	app := &application{
		config:  cfg,
		logger:  logger,
		catalog: catalogv1.NewCatalogServiceClient(catalogConn),
		booking: bookingv1.NewBookingServiceClient(bookingConn),
	}

	return app.serve()
}

func loadCfg() (config, bool) {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.catalogAddr, "catalog-addr", "localhost:50051", "Catalog service address")
	flag.StringVar(&cfg.bookingAddr, "booking-addr", "localhost:50052", "Booking service address")
	flag.DurationVar(&cfg.shutdownTimeout, "shutdown-timeout", 20*time.Second, "Maximum time to wait for in-flight work during shutdown")
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	return cfg, *displayVersion
}
