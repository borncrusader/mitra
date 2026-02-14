package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"mitra/internal/proto"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func Start(cfg *proto.Config) error {
	grpcServer := grpc.NewServer()
	repoService := NewRepoServiceServer()
	proto.RegisterRepoServiceServer(grpcServer, repoService)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.Server.GrpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Server.GrpcPort, err)
	}

	go func() {
		log.Printf("gRPC server starting on %s", cfg.Server.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	http.HandleFunc("/", helloHandler)
	log.Printf("HTTP server starting on http://localhost%s", cfg.Server.Port)

	return http.ListenAndServe(cfg.Server.Port, nil)
}
