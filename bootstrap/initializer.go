// Package bootstrap contains code used to configure the Raft process on startup.
package bootstrap

import (
	"github.com/jrobison153/raft/journal"
	"github.com/jrobison153/raft/policy/client"
	"github.com/jrobison153/raft/server/grpc"
	"github.com/jrobison153/raft/state"
)

func Init() *grpc.RaftServer {

	journal := journal.NewArrayJournal()
	renderer := state.NewMapStateMachine()

	clientPolicy := client.New(journal, renderer)
	server := grpc.New(clientPolicy)

	return server
}
