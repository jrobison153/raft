package main

import (
	"github.com/jrobison153/raft/bootstrap"
	"github.com/jrobison153/raft/environment"
)

const defaultPort = 3434

func main() {

	bootstrapper := bootstrap.New()

	err := bootstrapper.Init()

	if err != nil {
		panic(err)
	} else {

		resolvedPort, _ := environment.ResolvePort(defaultPort)

		bootstrapper.Start(resolvedPort)
	}
}
