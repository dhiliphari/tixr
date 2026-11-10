package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

func Run(ctx context.Context, logger *slog.Logger, srv *grpc.Server, hs *health.Server, addr string, shutdownTimeout time.Duration) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(lis)
	}()

	logger.Info("starting gRPC server", "addr", lis.Addr().String())

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	logger.Info("shutting down gRPC server", "timeout", shutdownTimeout.String())
	hs.Shutdown()

	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(shutdownTimeout)
	defer timer.Stop()

	select {
	case <-stopped:
		logger.Info("gRPC server stopped")
		return nil
	case <-timer.C:
		srv.Stop()
		<-stopped
		return errors.New("graceful stop timed out; in-flight RPCs were cancelled")
	}
}
