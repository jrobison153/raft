// Package client contains the Raft client interface core policy/business logic. This package
// forms the bridge between various possible controllers (gRPC, OpenAPI, TCP, etc.) that provide the client
// a physical means to interact with Raft
package client

import (
	"errors"
	"github.com/jrobison153/raft/journal"
	"github.com/jrobison153/raft/state"
	"reflect"
)

// Client is an instance of a Raft client with specific journal and replicator implementations
type Client struct {
	journal  journal.Journaler
	renderer state.Renderer
}

type Persister interface {
	Put(item []byte) (chan bool, error)
	Get(item []byte) ([]byte, error)
	TypeOfLogger() string
	TypeOfStateMachine() string
}

var (
	ErrKeyDoesNotExist                 = errors.New("attempt to get item for a key that does not exist")
	ErrEmptyKey                        = errors.New("attempt to put item with an empty key")
	ErrAppendToJournalFailed           = errors.New("failure appending entry to journal")
	ErrRegisterForNotificationOnCommit = errors.New("failure to register for notification on commit index")
)

// New Returns a newly initialized Client that will use journal for journaling
func New(journal journal.Journaler, renderer state.Renderer) *Client {

	return &Client{
		journal:  journal,
		renderer: renderer,
	}
}

// Put stores item in the journal and replicates to followers in the raft cluster.
// Put returns a bool channel that will signal when the replication has completed. The value true will
// be written to the channel should replication succeed and item has been committed to the journal. False
// will be written in the event replication and ultimately commit to the log fails.
// Put returns an error should there be any failure prior to attempting replication.
// ErrAppendToJournalFailed - failure to append the item to the journal
func (client *Client) Put(item []byte) (chan bool, error) {

	var putErr error
	var doneCh chan bool

	var index uint64
	index, putErr = client.appendToJournal(item)

	if putErr == nil {
		doneCh, putErr = registerForNotificationOnCommitIndex(client.journal, index)
	} else {
		doneCh = unblockedChannel(false)
	}

	return doneCh, putErr
}

// Get returns the item per the request item. An error is returned if the request is unable to be satisfied
func (client *Client) Get(item []byte) ([]byte, error) {

	data, err := client.renderer.ResolveRequestToData(item)

	var getErr error
	if err != nil {
		getErr = ErrKeyDoesNotExist
	}

	return data, getErr
}

func (client *Client) TypeOfLogger() string {

	return reflect.TypeOf(client.journal).String()
}

func (client *Client) TypeOfStateMachine() string {

	return reflect.TypeOf(client.renderer).String()
}

func unblockedChannel(val bool) chan bool {

	theCh := make(chan bool)
	go func(v bool) { theCh <- v }(val)

	return theCh
}

func (client *Client) appendToJournal(item []byte) (uint64, error) {

	result := client.journal.Append(item)

	appendResult := <-result

	var putErr error = nil

	if appendResult.Error != nil {
		putErr = ErrAppendToJournalFailed
	}

	return appendResult.Index, putErr
}

func registerForNotificationOnCommitIndex(journal journal.Journaler, index uint64) (chan bool, error) {

	doneCh := make(chan bool)

	notifyErr := journal.NotifyOfCommitOnIndexOnce(index, doneCh)

	var err error

	if notifyErr != nil {
		// TODO need telemetry here
		doneCh = unblockedChannel(false)
		err = ErrRegisterForNotificationOnCommit
	}

	return doneCh, err
}
