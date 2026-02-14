package server

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"mitra/internal/config"
	"mitra/internal/proto"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func Start(cfg *config.Config) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Logger()

	grpcServer := grpc.NewServer()
	repoService := NewRepoServiceServer(logger)
	proto.RegisterRepoServiceServer(grpcServer, repoService)
	reflection.Register(grpcServer)

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

	http.HandleFunc("/", helloHandler)
	logger.Info().
		Str("port", cfg.Server.Port).
		Msgf("HTTP server starting")

	return http.ListenAndServe(cfg.Server.Port, nil)
}
