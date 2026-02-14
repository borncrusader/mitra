package server

import (
	"context"

	"mitra/internal/proto"
)

type RepoServiceServer struct {
	proto.UnimplementedRepoServiceServer
}

func NewRepoServiceServer() *RepoServiceServer {
	return &RepoServiceServer{}
}

func (s *RepoServiceServer) ListRepos(ctx context.Context, req *proto.ListReposRequest) (*proto.ListReposResponse, error) {
	return &proto.ListReposResponse{
		Repos: []*proto.Repo{},
	}, nil
}
