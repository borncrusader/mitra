package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/util"
)

type ClientContext struct {
	Client proto.MitraServiceClient
	Ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

func (c *ClientContext) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		util.DeferCheck(c.conn.Close)
	}
}

func newClient() (*ClientContext, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := proto.NewMitraServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	return &ClientContext{
		Client: client,
		Ctx:    ctx,
		cancel: cancel,
		conn:   conn,
	}, nil
}
