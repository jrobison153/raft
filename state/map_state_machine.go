package state

import (
	"encoding/json"
	"errors"
	"github.com/jrobison153/raft/journal"
	"log"
	"reflect"
)

type MapStateMachine struct {
	journal                journal.Journaler
	commitListenCh         chan uint64
	data                   map[string][]byte
	highestSeenCommitIndex int64
	isRunning              bool
}

// KeyValItem TODO - this struct is the same as the client "drivers" type, can we share it?
type KeyValItem struct {
	Key  string
	Data []byte
}

// Key TODO - this struct is the same as the client "drivers" type, can we share it?
type Key struct {
	Key string
}

var (
	ErrAlreadyRunning            = errors.New("state machine already running")
	ErrCannotUnmarshalKeyValData = errors.New("data in journal is not correct key value format")
	ErrInvalidKeyRequest         = errors.New("attempt to get value with a request that is not valid Key type")
	ErrKeyNotFound               = errors.New("key not associated with any data in store")
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
// Returns ErrCannotUnmarshalKeyValData if there is are failures trying to load state machine with key value data
// from the journal
func (state *MapStateMachine) Start() error {

	var err error

	if !state.isRunning {

		err = state.renderCurrentStateFromJournal()

		if err == nil {

			state.commitListenCh = make(chan uint64)
			state.journal.NotifyOfAllCommitChanges(state.commitListenCh)

			go state.listenForStateChanges(state.commitListenCh)

			state.isRunning = true
		}
	} else {
		err = ErrAlreadyRunning
	}

	return err
}

// ResolveRequestToData returns the value associated with request.
// Returns ErrKeyNotfound If the key is not found in the data store
func (state *MapStateMachine) ResolveRequestToData(request []byte) ([]byte, error) {

	key := &Key{}

	marshalErr := json.Unmarshal(request, key)

	var err error
	var val []byte

	if marshalErr == nil {

		var ok bool
		val, ok = state.data[key.Key]

		if !ok {
			err = ErrKeyNotFound
		}
	} else {
		log.Printf("Invalid request '%v', must be type Key with the following structure %+v", request, Key{})
		err = ErrInvalidKeyRequest
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

			// explicitly ignoring error response here, nothing we can do if the data
			// in the journal is corrupt or invalid. Error will be logged
			//nolint:errcheck
			state.updateStateWithJournalItem(journalEntry.Item)
		}

		state.highestSeenCommitIndex = int64(index)
	}
}

func (state *MapStateMachine) renderCurrentStateFromJournal() error {

	committedEntriesIt := state.journal.GetAllCommittedEntries()
	var err error

	for committedEntriesIt.HasNext() && err == nil {

		committedEntry, _ := committedEntriesIt.Next()

		rawKeyVal := committedEntry.Item

		err = state.updateStateWithJournalItem(rawKeyVal)
	}

	return err
}

func (state *MapStateMachine) updateStateWithJournalItem(rawKeyVal []byte) error {

	keyValItem := &KeyValItem{}

	unMarshalErr := json.Unmarshal(rawKeyVal, keyValItem)
	var err error

	if unMarshalErr != nil {
		log.Printf("unable to update state machine, data in journal not correct k/v format")
		err = ErrCannotUnmarshalKeyValData
	} else {
		state.data[keyValItem.Key] = keyValItem.Data
	}

	return err
}
