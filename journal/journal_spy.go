package journal

import (
	"errors"
	"reflect"
)

type journalEntry struct {
	key  []byte
	data []byte
}

type Spy struct {
	log                                    []journalEntry
	appendCalled                           bool
	spiedAppendKey                         []byte
	spiedAppendData                        []byte
	isFailingNextAppend                    bool
	appendFailMsg                          string
	subscribers                            map[uint64]chan bool
	isFailingNextNotifyOfCommitOnIndexOnce bool
}

func (spy *Spy) GetHead() (Entry, error) {
	//TODO implement me
	panic("implement me")
}

func New() *Spy {

	return &Spy{
		log:                                    make([]journalEntry, 0, 16),
		appendCalled:                           false,
		isFailingNextAppend:                    false,
		isFailingNextNotifyOfCommitOnIndexOnce: false,
		subscribers:                            make(map[uint64]chan bool),
	}
}

// Begin Journaler interface

func (spy *Spy) Append(key []byte, data []byte) (uint64, error) {

	var err error

	if spy.isFailingNextAppend {
		err = errors.New(spy.appendFailMsg)
	} else {

		spy.appendCalled = true
		spy.spiedAppendKey = key
		spy.spiedAppendData = data

		entry := journalEntry{
			key:  key,
			data: data,
		}

		spy.log = append(spy.log, entry)

		err = nil
	}

	headIndex := len(spy.log) - 1

	return uint64(headIndex), err
}

func (spy *Spy) NotifyOfCommitOnIndexOnce(index uint64, ch chan bool) error {

	var err error
	if spy.isFailingNextNotifyOfCommitOnIndexOnce {
		err = errors.New("failing NotifyOfCommitOnIndexOnce for test purposes")
	} else {
		spy.subscribers[index] = ch
	}

	return err
}

func (spy *Spy) Commit(index uint64) error {

	notifySubscribersOfIndexChange(index, spy.subscribers)
	return nil
}

func notifySubscribersOfIndexChange(currentCommitIndex uint64, subscribers map[uint64]chan bool) {

	for commitIndex, ch := range subscribers {

		if currentCommitIndex >= commitIndex {
			ch <- true
		}
	}
}

// End Journaler interface

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

func (spy *Spy) RegisteredForNotifyOnIndex(journalIndex uint64) bool {

	return spy.subscribers[journalIndex] != nil
}

func (spy *Spy) RegisteredNotifyChannelOnIndex(journalIndex uint64, ch chan bool) bool {

	return ch == spy.subscribers[journalIndex]
}

func (spy *Spy) FailNextNotifyOfCommitOnIndexOnce() {
	spy.isFailingNextNotifyOfCommitOnIndexOnce = true
}

// End Spy functions
