// Package client contains the Raft client interface core policy/business logic. This package
// forms the bridge between various possible controllers (gRPC, OpenAPI, TCP, etc.) that provide the client
// a physical means to interact with Raft
package client

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/jrobison153/raft/journal"
	"reflect"
)

// Client is an instance of a Raft client with specific journal and replicator implementations
type Client struct {
	journal journal.Journaler
}

type Persister interface {
	Put(key string, data []byte) (chan bool, error)
	Get(key string) ([]byte, error)
	TypeOfLogger() string
}

var (
	ErrKeyDoesNotExist   = errors.New("attempt to get data for a key that does not exist")
	ErrEmptyKey          = errors.New("attempt to put data with an empty key")
	ErrAppendToLogFailed = errors.New("failure appending entry to journal")
)

func (client *Client) TypeOfLogger() string {

	return reflect.TypeOf(client.journal).String()
}

// New Returns a newly initialized Client that will use journal for journaling
func New(journal journal.Journaler) *Client {

	return &Client{
		journal: journal,
	}
}

// Put stores data associated with key replicating the both to followers in the raft cluster.
// Put returns a bool channel that will signal when the replication has completed. The value true will
// be written to the channel should replication succeed and data has been committed to the journal. False
// will be written in the event replication and ultimately commit to the log fails.
// Put returns an error should there be any failure prior to attempting replication.
func (client *Client) Put(key string, data []byte) (chan bool, error) {

	var putErr error
	var doneCh chan bool

	if len(key) == 0 {

		putErr = ErrEmptyKey
		doneCh = unblockedChannel(false)
	} else {

		hashedKey := hashKey(key)

		var index uint64
		index, putErr = appendToJournal(hashedKey, data, client.journal)

		if putErr == nil {

			doneCh = registerForNotificationOnCommitIndex(client.journal, index)
		} else {

			putErr = ErrAppendToLogFailed
			doneCh = unblockedChannel(false)
		}
	}

	return doneCh, putErr
}

func registerForNotificationOnCommitIndex(journal journal.Journaler, index uint64) chan bool {

	doneCh := make(chan bool)
	notifyErr := journal.NotifyOfCommitOnIndexOnce(index, doneCh)

	if notifyErr != nil {
		doneCh = unblockedChannel(false)
	}

	return doneCh
}

// Get returns the data associated with Key. An error is returned if Key is not associated with any
// data, i.e. a previously Put item
func (client *Client) Get(key string) ([]byte, error) {

	return []byte{}, ErrKeyDoesNotExist
}

func hashKey(key string) []byte {

	hashFunc := sha1.New()

	hashedKey := hashFunc.Sum([]byte(key))

	return hashedKey
}

func unblockedChannel(val bool) chan bool {

	theCh := make(chan bool)
	go func(v bool) { theCh <- v }(val)

	return theCh
}

func appendToJournal(key []byte, data []byte, log journal.Journaler) (uint64, error) {

	index, err := log.Append(key, data)

	var putErr error = nil

	if err != nil {
		putErr = fmt.Errorf("client: append to journal failed, root cause '%s'", err.Error())
	}

	return index, putErr
}
