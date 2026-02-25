package cli

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/proto"
	"mitra/internal/util"
)

type clientContext struct {
	Client proto.MitraServiceClient
	Ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

func (cc *clientContext) Close() {
	if cc.cancel != nil {
		cc.cancel()
	}
	if cc.conn != nil {
		util.DeferCheck(cc.conn.Close)
	}
}

func (c *Command) newClient() (*clientContext, error) {
	conn, err := grpc.NewClient("localhost"+c.Cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := proto.NewMitraServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	return &clientContext{
		Client: client,
		Ctx:    ctx,
		cancel: cancel,
		conn:   conn,
	}, nil
}
