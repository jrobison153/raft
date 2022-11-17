package main

import (
	"github.com/jrobison153/raft/bootstrap"
	"github.com/jrobison153/raft/environment"
)

const defaultPort = 3434

func main() {

	server := bootstrap.Init()

	resolvedPort, _ := environment.ResolvePort(defaultPort)

	server.Start(resolvedPort)
}
