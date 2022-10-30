package server

import (
	"context"
	"fmt"
	"github.com/jrobison153/raft/api"
	"google.golang.org/grpc"
	"log"
	"net"
)

type RaftClientApi struct {
	api.UnimplementedPersisterServer
	server *grpc.Server
}

func (server *RaftClientApi) PutItem(ctx context.Context, item *api.Item) (*api.PutResponse, error) {
	return &api.PutResponse{}, nil
}

func (server *RaftClientApi) mustEmbedUnimplementedPersisterServer() {
	//TODO implement me
	panic("implement me")
}

func (server *RaftClientApi) Start(port uint32) {

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	server.server = grpcServer

	api.RegisterPersisterServer(grpcServer, server)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("failed to start server %s", err)
	} else {
		log.Printf("server started on port %d\n", port)
	}
}

func (server *RaftClientApi) Stop() {

	server.server.Stop()
	log.Println("server stopped")
}
