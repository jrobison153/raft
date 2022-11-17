// Package client contains the Raft client interface core policy/business logic. This package
// forms the bridge between various possible controllers (gRPC, OpenAPI, TCP, etc.) that provide the client
// a physical means to interact with Raft
package client

import (
	"crypto/sha1"
	"errors"
	"fmt"
	logger "github.com/jrobison153/raft/log"
	"github.com/jrobison153/raft/replicator"
	"reflect"
)

// Client is an instance of a Raft client with specific log and replicator implementations
type Client struct {
	log        logger.Logger
	replicator replicator.Replicator
}

type Persister interface {
	Put(key string, data []byte) (chan bool, error)
	Get(key string) ([]byte, error)
	TypeOfLogger() string
	TypeOfReplicator() string
}

var (
	ErrKeyDoesNotExist = errors.New("attempt to get data for a key that does not exist")
	ErrEmptyKey        = errors.New("attempt to put data with an empty key")
)

func (client *Client) TypeOfReplicator() string {

	return reflect.TypeOf(client.replicator).String()
}

func (client *Client) TypeOfLogger() string {

	return reflect.TypeOf(client.log).String()
}

// New Returns a newly initialized Client that will use log for journaling and replicator
// to manage replication to followers in the Raft cluster
func New(log logger.Logger, replicator replicator.Replicator) *Client {

	return &Client{
		log:        log,
		replicator: replicator,
	}
}

// Put stores data associated with key replicating the both to followers in the raft cluster.
// Put returns a bool channel that will signal when the replication has completed. The value true will
// be written to the channel should replication succeed and false will be written in the event replication
// fails. Put returns an error should there be any failure prior to attempting replication.
func (client *Client) Put(key string, data []byte) (chan bool, error) {

	var putErr error
	var doneCh chan bool

	if len(key) == 0 {

		putErr = ErrEmptyKey
		doneCh = unblockedChannel()
	} else {

		hashedKey := hashKey(key)

		putErr = appendToLog(hashedKey, data, client.log)

		if putErr == nil {
			doneCh = replicateData(hashedKey, data, client.replicator)
		} else {
			doneCh = unblockedChannel()
		}
	}

	return doneCh, putErr
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

func unblockedChannel() chan bool {

	theCh := make(chan bool)
	go func() { theCh <- false }()

	return theCh
}

func replicateData(key []byte, data []byte, repl replicator.Replicator) chan bool {

	doneCh := make(chan bool)

	item := &replicator.Item{
		Key:  key,
		Data: data,
	}

	go repl.Replicate(item, doneCh)

	return doneCh
}

func appendToLog(key []byte, data []byte, log logger.Logger) error {

	err := log.Append(key, data)

	var putErr error = nil

	if err != nil {
		putErr = fmt.Errorf("client: append to log failed, root cause '%s'", err.Error())
	}

	return putErr
}
