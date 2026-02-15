package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/util"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func Start(cfg *config.Config) error {
	logger := util.NewLogger(os.Stdout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	grpcServer := grpc.NewServer()
	repoService, err := NewRepoServiceServer(logger, cfg, ctx, &wg)
	if err != nil {
		return fmt.Errorf("failed to create repo service: %w", err)
	}
	proto.RegisterRepoServiceServer(grpcServer, repoService)
	reflection.Register(grpcServer)

	if err := repoService.StartExistingWatchers(); err != nil {
		logger.Error().Err(err).Msg("failed to start existing watchers")
	}

	lis, err := net.Listen("tcp", cfg.Server.GrpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Server.GrpcPort, err)
	}

	go func() {
		logger.Info().
			Str("port", cfg.Server.GrpcPort).
			Msg("gRPC server starting")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal().
				Err(err).
				Msg("failed to serve gRPC")
		}
	}()

	httpServer := &http.Server{
		Addr: cfg.Server.Port,
	}
	http.HandleFunc("/", helloHandler)

	go func() {
		logger.Info().
			Str("port", cfg.Server.Port).
			Msg("HTTP server starting")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().
				Err(err).
				Msg("HTTP server error")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info().Msg("shutdown signal received, initiating graceful shutdown")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("HTTP server shutdown error")
	}

	grpcServer.GracefulStop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info().Msg("all watchers stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn().Msg("shutdown timeout exceeded, some watchers may not have stopped")
	}

	logger.Info().Msg("server shutdown complete")
	return nil
}
