// Package bootstrap contains code used to configure the Raft process on startup.
package bootstrap

import (
	"errors"
	"github.com/jrobison153/raft/journal"
	"github.com/jrobison153/raft/policy/client"
	"github.com/jrobison153/raft/replication"
	"github.com/jrobison153/raft/server"
	"github.com/jrobison153/raft/server/grpc"
	"github.com/jrobison153/raft/state"
	"log"
	"os"
	"strings"
)

type Bootstrap struct {
	journal      journal.Journaler
	stateMachine state.Renderer
	server       server.LifeCycler
	clientPolicy client.Persister
	replicator   replication.Replicator
}

const (
	journalTypeEnvVar      = "JOURNAL_TYPE"
	replicatorTypeEnvVar   = "REPLICATOR_TYPE"
	stateMachineTypeEnvVar = "STATE_MACHINE_TYPE"

	spyJournalType      = "SPY"
	spyReplicatorType   = "SPY"
	spyStateMachineType = "SPY"
)

var (
	ErrInvalidJournalType      = errors.New("journal type specified in the environment is not supported")
	ErrInvalidReplicatorType   = errors.New("replicator type specified in the environment is not supported")
	ErrInvalidStateMachineType = errors.New("state machine type specified in the environment is not supported")
)

func New() *Bootstrap {

	return &Bootstrap{}
}

// Init creates the correct implementations of the journal, stateMachine, and replicator. Initialization
// does not involve starting any components, i.e. this is just configuration and setup. Method Start
// can be used to bring the components up in the required correct order
func (bootstrapper *Bootstrap) Init() error {

	var err error
	bootstrapper.journal, err = resolveJournalImpl()

	if err == nil {

		err = bootstrapper.initializeStateMachineAndReplicator()

		if err == nil {

			bootstrapper.clientPolicy = client.New(bootstrapper.journal, bootstrapper.stateMachine)
			bootstrapper.server = grpc.New(bootstrapper.clientPolicy)
		}
	}

	return err
}

// Start starts the replicator, the state machine (listening for changes to the journal commit index),
// and the API servers, client APIs listening on port clientApiServerPort
func (bootstrapper *Bootstrap) Start(clientApiServerPort uint32) {

	// TODO handle error on state machine start up
	_ = bootstrapper.stateMachine.Start()

	timer := replication.NewSleepTimer()
	bootstrapper.replicator.Start(timer)

	bootstrapper.server.Start(clientApiServerPort)
}

func (bootstrapper *Bootstrap) initializeStateMachineAndReplicator() error {

	var err error
	var stateMachineResolveErr error
	var replResolveErr error

	bootstrapper.replicator, replResolveErr = resolveReplicator(bootstrapper.journal)

	bootstrapper.stateMachine, stateMachineResolveErr = resolveStateMachineImpl(bootstrapper.journal)

	if replResolveErr != nil {
		err = replResolveErr
	} else if stateMachineResolveErr != nil {
		err = stateMachineResolveErr
	}

	return err
}

func resolveStateMachineImpl(journal journal.Journaler) (state.Renderer, error) {

	stateMachineType, isStateMachineTypeSet := os.LookupEnv(stateMachineTypeEnvVar)

	var theStateMachine state.Renderer
	var err error

	if isStateMachineTypeSet {

		if strings.Compare(stateMachineType, spyStateMachineType) == 0 {
			theStateMachine = state.NewStateMachineSpy(journal)
		} else {
			err = ErrInvalidStateMachineType
			log.Printf("Unknown state machine type '%s'", stateMachineType)
		}
	} else {
		theStateMachine = state.NewMapStateMachine(journal)
	}

	return theStateMachine, err
}

func resolveJournalImpl() (journal.Journaler, error) {

	journalType, isJournalTypeSet := os.LookupEnv(journalTypeEnvVar)

	var theJournal journal.Journaler
	var err error

	if isJournalTypeSet {

		if strings.Compare(journalType, spyJournalType) == 0 {
			theJournal = journal.NewJournalSpy()
		} else {
			err = ErrInvalidJournalType
			log.Printf("Unknown journal type '%s'", journalType)
		}
	} else {
		theJournal = journal.NewArrayJournal()
	}

	return theJournal, err
}

func resolveReplicator(journal journal.Journaler) (replication.Replicator, error) {

	replicatorType, isReplicatorTypeSet := os.LookupEnv(replicatorTypeEnvVar)

	var theReplicator replication.Replicator

	var err error

	if isReplicatorTypeSet {

		if strings.Compare(replicatorType, spyReplicatorType) == 0 {
			theReplicator = replication.NewReplicatorSpy(journal)
		} else {

			err = ErrInvalidReplicatorType
			log.Printf("unknown replicator type '%s'", replicatorType)
		}
	} else {
		config := replication.NewDefaultConfig()
		theReplicator = replication.NewNoOpReplicator(journal, config)
	}

	return theReplicator, err
}
