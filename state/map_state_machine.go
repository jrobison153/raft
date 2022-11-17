package state

import (
	"errors"
	"github.com/jrobison153/raft/journal"
	"reflect"
)

type MapStateMachine struct {
	journal                journal.Journaler
	commitListenCh         chan uint64
	data                   map[string][]byte
	highestSeenCommitIndex int64
	isRunning              bool
}

var (
	ErrAlreadyRunning = errors.New("state machine already running")
	ErrKeyNotFound    = errors.New("key not associated with any data in store")
)

func NewMapStateMachine(journal journal.Journaler) *MapStateMachine {

	return &MapStateMachine{
		journal:                journal,
		data:                   make(map[string][]byte),
		highestSeenCommitIndex: -1,
		isRunning:              false,
	}
}

// Start registers this state machine as a listener on all commit changes of the journal.
// Returns ErrAlreadyRunning if this method is called more than once
func (state *MapStateMachine) Start() error {

	var err error

	if !state.isRunning {

		committedEntriesIt := state.journal.GetAllCommittedEntries()

		for committedEntriesIt.HasNext() {

			committedEntry, _ := committedEntriesIt.Next()

			state.data[committedEntry.RawKey] = committedEntry.Data
		}

		state.commitListenCh = make(chan uint64)
		state.journal.NotifyOfAllCommitChanges(state.commitListenCh)

		go state.listenForStateChanges(state.commitListenCh)

		state.isRunning = true
	} else {
		err = ErrAlreadyRunning
	}

	return err
}

// GetValueForKey returns the value associated with key. If the key is not found in the
// data store then ErrKeyNotfound is returned
func (state *MapStateMachine) GetValueForKey(key string) ([]byte, error) {

	val, ok := state.data[key]

	var err error

	if !ok {
		err = ErrKeyNotFound
	}

	return val, err
}

func (state *MapStateMachine) TypeOfLogger() string {

	return reflect.TypeOf(state.journal).String()
}

func (state *MapStateMachine) listenForStateChanges(ch chan uint64) {

	var index uint64
	for {

		index = <-ch

		startIndex := uint64(state.highestSeenCommitIndex + 1)
		entries, _ := state.journal.GetAllEntriesBetween(startIndex, index)

		for entries.HasNext() {

			journalEntry, _ := entries.Next()
			state.data[journalEntry.RawKey] = journalEntry.Data
		}

		state.highestSeenCommitIndex = int64(index)
	}
}
