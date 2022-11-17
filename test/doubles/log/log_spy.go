package log

import (
	"errors"
	"github.com/jrobison153/raft/log"
	"reflect"
)

type logEntry struct {
	key  []byte
	data []byte
}

type Spy struct {
	log                 []logEntry
	appendCalled        bool
	spiedAppendKey      []byte
	spiedAppendData     []byte
	isFailingNextAppend bool
	appendFailMsg       string
}

func (spy *Spy) GetHead() (log.Entry, error) {
	//TODO implement me
	panic("implement me")
}

func New() *Spy {

	return &Spy{
		log:                 make([]logEntry, 16),
		appendCalled:        false,
		isFailingNextAppend: false,
	}
}

// Begin Logger interface

func (spy *Spy) Append(key []byte, data []byte) error {

	var result error

	if spy.isFailingNextAppend {
		result = errors.New(spy.appendFailMsg)
	} else {

		spy.appendCalled = true
		spy.spiedAppendKey = key
		spy.spiedAppendData = data

		entry := logEntry{
			key:  key,
			data: data,
		}

		spy.log = append(spy.log, entry)

		result = nil
	}

	return result
}

// End Logger interface

// Begin Spy Functions

func (spy *Spy) AppendCalled() bool {

	return spy.appendCalled
}

func (spy *Spy) AppendKey() []byte {

	return spy.spiedAppendKey
}

func (spy *Spy) AppendData() []byte {

	return spy.spiedAppendData
}

func (spy *Spy) LatestDataForKey(key []byte) []byte {

	var mostRecentAppendForKey []byte

	for _, entry := range spy.log {

		if reflect.DeepEqual(entry.key, key) {
			mostRecentAppendForKey = entry.data
			break
		}
	}

	return mostRecentAppendForKey
}

func (spy *Spy) FailNextAppend(appendFailMsg string) {

	spy.isFailingNextAppend = true
	spy.appendFailMsg = appendFailMsg
}

// End Spy functions
