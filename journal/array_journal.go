package journal

import (
	"errors"
	"sync"
)

// ArrayJournal implements an append only Last In First Out (LIFO) stack. This type allows for safe concurrent
// usage as it is intended to be used as a singleton in a highly parallel execution environment
type ArrayJournal struct {
	theLog      []Entry              // access to this value must be synchronized
	headIndex   int64                // access to this value must be synchronized
	subscribers map[uint64]chan bool // access to this value must be synchronized
	locker      sync.Locker
	commitIndex uint64 // access to this value must be synchronized
}

var (
	ErrCommitOnEmptyLog       = errors.New("attempt to commit on an empty log")
	ErrEmptyLog               = errors.New("attempt to get Head item from an empty log")
	ErrIndexBeyondHead        = errors.New("attempt to commit an index that is beyond the head of the log")
	ErrSubscriptionOnEmptyLog = errors.New("attempt to setup subscription to index on an empty log")
)

func NewArrayJournal() *ArrayJournal {

	return &ArrayJournal{
		theLog:      make([]Entry, 0, 1024),
		headIndex:   -1,
		subscribers: make(map[uint64]chan bool),
		locker:      &sync.Mutex{},
	}
}

// Append adds an Entry to the log as the new Head item with key and data set as the respective Entry
// data elements.
// Append is safe for concurrent execution
func (logger *ArrayJournal) Append(key []byte, data []byte) (uint64, error) {

	entry := Entry{
		Key:  key,
		Data: data,
	}

	logger.locker.Lock()
	defer logger.locker.Unlock()

	logger.theLog = append(logger.theLog, entry)
	logger.headIndex += 1

	return uint64(logger.headIndex), nil
}

// GetHead returns the head Entry of the log, i.e. the last Entry that was added via Append.
// If the log is empty then an ErrEmptyLog is returned.
// GetHead is safe for concurrent execution
func (logger *ArrayJournal) GetHead() (Entry, error) {

	var headEntry Entry
	var err error

	logger.locker.Lock()
	defer logger.locker.Unlock()

	if logger.headIndex >= 0 {

		headEntry = logger.theLog[logger.headIndex]
	} else {
		err = ErrEmptyLog
	}

	return headEntry, err
}

// NotifyOfCommitOnIndexOnce registers notification channel ch with log index. When a Commit is made
// on an index greater than or equal to index, then the registered ch will receive a value of true if there are
// no errors and false if there was an error.
// NotifyOfCommitOnIndexOnce is safe for concurrent execution
func (logger *ArrayJournal) NotifyOfCommitOnIndexOnce(index uint64, ch chan bool) error {

	var err error

	logger.locker.Lock()
	defer logger.locker.Unlock()

	if logger.isEmptyLog() {

		err = ErrSubscriptionOnEmptyLog
	} else if logger.isCommitIndexWithinValidRange(index) {

		logger.subscribers[index] = ch
	} else {

		err = ErrIndexBeyondHead
	}

	return err
}

// Commit attempts to commit the index. Per the Raft protocol once an index is committed then all
// previous uncommitted log entries are also committed. On successful Commit all subscribers registered
// via NotifyOfCommitOnIndexOnce are notified via their registered channel and then removed. Subscribed
// channels receive the value true. If an attempt is made to
// Commit an index that is beyond the head of the log, i.e. committing an index that is greater than the underlying
// number of log entries, then the error ErrIndexBeyondHead is returned. If an attempt is made to Commit an index on
// an empty log then the ErrCommitOnEmptyLog error is returned
// Commit is safe for concurrent execution
func (logger *ArrayJournal) Commit(index uint64) error {

	var err error

	logger.locker.Lock()
	defer logger.locker.Unlock()

	if logger.isEmptyLog() {
		err = ErrCommitOnEmptyLog
	} else if logger.isCommitIndexWithinValidRange(index) {

		logger.notifySubscribers(index)
	} else {
		err = ErrIndexBeyondHead
	}

	return err
}

func (logger *ArrayJournal) notifySubscribers(index uint64) {

	logger.commitIndex = index

	for subscribedIndex, notificationCh := range logger.subscribers {

		if logger.commitIndex >= subscribedIndex {

			// must delete subscriber before sending notification to ensure cleanup is done
			// before subscriber is notified
			delete(logger.subscribers, subscribedIndex)
			notificationCh <- true
		}
	}
}

func (logger *ArrayJournal) SetMemoryLocker(locker sync.Locker) {
	logger.locker = locker
}

func (logger *ArrayJournal) isCommitIndexWithinValidRange(index uint64) bool {

	return index <= uint64(logger.headIndex)
}

func (logger *ArrayJournal) isEmptyLog() bool {

	return logger.headIndex == -1
}
