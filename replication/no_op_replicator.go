package replication

import (
	"github.com/jrobison153/raft/journal"
	"reflect"
)

type NoOpReplicator struct {
	journal journal.Journaler
	config  *Config
	timer   Timer
}

func NewNoOpReplicator(journal journal.Journaler, config *Config) *NoOpReplicator {

	return &NoOpReplicator{
		journal: journal,
		config:  config,
	}
}

func (repl *NoOpReplicator) Start(timer Timer) {

	repl.timer = timer
	go repl.replicateCommits()
}

func (repl *NoOpReplicator) TypeOfLogger() string {

	return reflect.TypeOf(repl.journal).String()
}

func (repl *NoOpReplicator) replicateCommits() {

	for {

		uncommittedEntriesResultCh := repl.journal.GetAllUncommittedEntries()

		committedEntriesResult := <-uncommittedEntriesResultCh

		if committedEntriesResult.HasUncommittedEntries {

			commitResultCh := repl.journal.Commit(uint64(committedEntriesResult.HeadIndex))

			<-commitResultCh
		}

		repl.timer.WaitMs(repl.config.JournalPollPeriod)
	}

}

func (repl *NoOpReplicator) ReplicatorConfig() Config {

	return *(repl.config)
}
