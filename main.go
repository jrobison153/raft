package main

import (
	"github.com/jrobison153/raft/environment"
	"github.com/jrobison153/raft/server/grpc"
)

const defaultPort = 3434

func main() {

	server := grpc.New()

	resolvedPort, _ := environment.ResolvePort(defaultPort)

	server.Start(resolvedPort)
}
