package journal

import (
	"errors"
)

var (
	ErrCommitOnEmptyLog       = errors.New("attempt to commit on an empty log")
	ErrEmptyLog               = errors.New("attempt to get Head item from an empty log")
	ErrIndexBeyondHead        = errors.New("attempt to commit an index that is beyond the head of the log")
	ErrIndexOutOfBounds       = errors.New("attempt to access journal item with an out of bounds index")
	ErrInvertedIndexes        = errors.New("start index is greater than end index, indexes are inverted")
	ErrSubscriptionOnEmptyLog = errors.New("attempt to setup subscription to index on an empty log")
)

const (
	Append                   = "append"
	Commit                   = "commit"
	GetAllUncommittedEntries = "get-all-uncommitted-entries"
)

type ArrayJournalCommand struct {
	appendDoneCh                   chan AppendResult
	appendEntry                    Entry
	commitDoneCh                   chan CommitResult
	commitIndex                    int64
	getAllUncommittedEntriesDoneCh chan AllUncommittedEntriesResult
	name                           string
}

// ArrayJournal implements an append only Last In First Out (LIFO) stack. This type allows for safe concurrent
// usage as it is intended to be used as a singleton in a highly parallel execution environment
type ArrayJournal struct {
	theLog                             []Entry
	headIndex                          int64
	oneTimeCommitChangeSubscribers     map[uint64][]chan bool
	commitIndex                        int64
	workQueue                          chan ArrayJournalCommand
	notifyOfAllCommitChangeSubscribers []chan uint64
}

func NewArrayJournal() *ArrayJournal {

	journal := &ArrayJournal{
		workQueue:                          make(chan ArrayJournalCommand, 1024),
		theLog:                             make([]Entry, 0, 1024),
		headIndex:                          -1,
		commitIndex:                        -1,
		oneTimeCommitChangeSubscribers:     make(map[uint64][]chan bool),
		notifyOfAllCommitChangeSubscribers: make([]chan uint64, 0, 16),
	}

	go journal.processCommands()

	return journal
}

// Append adds an Entry to the log as the new Head item with Key and Item set as the respective Entry
// Item elements.
// Append is safe for concurrent execution
func (journal *ArrayJournal) Append(item []byte) chan AppendResult {

	entry := Entry{
		Item: item,
	}

	doneCh := make(chan AppendResult)

	appendItem := ArrayJournalCommand{
		name:         Append,
		appendEntry:  entry,
		appendDoneCh: doneCh,
	}

	journal.workQueue <- appendItem

	return doneCh
}

// GetHead returns the head Entry of the log, i.e. the last Entry that was added via Append.
// If the log is empty then an ErrEmptyLog is returned.
// GetHead is safe for concurrent execution
func (journal *ArrayJournal) GetHead() (Entry, error) {

	var headEntry Entry
	var err error

	if journal.headIndex >= 0 {

		headEntry = journal.theLog[journal.headIndex]
	} else {
		err = ErrEmptyLog
	}

	return headEntry, err
}

// NotifyOfCommitOnIndexOnce registers notification channel ch with log index. When a Commit is made
// on an index greater than or equal to index, then the registered ch will receive a value of true if there are
// no errors and false if there was an error.
// NotifyOfCommitOnIndexOnce is safe for concurrent execution
func (journal *ArrayJournal) NotifyOfCommitOnIndexOnce(index uint64, ch chan bool) error {

	var err error

	// TODO need telemetry here
	if journal.isEmptyLog() {

		err = ErrSubscriptionOnEmptyLog
	} else if journal.isCommitIndexWithinValidRange(index) {

		journal.oneTimeCommitChangeSubscribers[index] = append(journal.oneTimeCommitChangeSubscribers[index], ch)
	} else {

		err = ErrIndexBeyondHead
	}

	return err
}

// Commit attempts to commit the index. Per the Raft protocol once an index is committed then all
// previous uncommitted log entries are also committed. On successful Commit all oneTimeCommitChangeSubscribers registered
// via NotifyOfCommitOnIndexOnce are notified via their registered channel and then removed. Subscribed
// channels receive the value true. On successful Commit all NotifyOfAlCommitChanges subscribers are notified of the
// change.
// If an attempt is made to
// Commit an index that is beyond the head of the log, i.e. committing an index that is greater than the underlying
// number of log entries, then the error ErrIndexBeyondHead is returned. If an attempt is made to Commit an index on
// an empty log then the ErrCommitOnEmptyLog error is returned
// Commit is safe for concurrent execution
func (journal *ArrayJournal) Commit(index uint64) chan CommitResult {

	doneCh := make(chan CommitResult)

	command := ArrayJournalCommand{
		name:         Commit,
		commitIndex:  int64(index),
		commitDoneCh: doneCh,
	}

	journal.workQueue <- command

	return doneCh
}

// GetAllCommittedEntries returns a read only Iterator wrapping the committed entries in the journal
func (journal *ArrayJournal) GetAllCommittedEntries() Iterator {

	committedEntries := journal.theLog[0 : journal.commitIndex+1]

	return NewArrayJournalIterator(committedEntries)
}

// GetAllEntriesBetween returns a read only Iterator wrapping journal entries between beginIndex and sizeOfBackingArray inclusive
// The values returned can include those that have not been committed yet.
func (journal *ArrayJournal) GetAllEntriesBetween(beginIndex uint64, endIndex uint64) (Iterator, error) {

	var iterator Iterator

	err := journal.validateIndexes(beginIndex, endIndex)

	if err == nil {
		entries := journal.theLog[beginIndex : endIndex+1]

		iterator = NewArrayJournalIterator(entries)
	}

	return iterator, err
}

func (journal *ArrayJournal) GetAllUncommittedEntries() chan AllUncommittedEntriesResult {

	doneCh := make(chan AllUncommittedEntriesResult)

	command := ArrayJournalCommand{
		name:                           GetAllUncommittedEntries,
		getAllUncommittedEntriesDoneCh: doneCh,
	}

	journal.workQueue <- command

	return doneCh
}

// NotifyOfAllCommitChanges registers channel ch to receive notifications when indexes in the journal
// are committed. The channel ch will receive the index that was committed.
func (journal *ArrayJournal) NotifyOfAllCommitChanges(ch chan uint64) {

	journal.notifyOfAllCommitChangeSubscribers = append(journal.notifyOfAllCommitChangeSubscribers, ch)
}

func (journal *ArrayJournal) notifySubscribers(commitIndex uint64) {

	journal.notifyOneTimeSubscribers(commitIndex)

	journal.notifyDurableSubscribers(commitIndex)
}

func (journal *ArrayJournal) notifyDurableSubscribers(commitIndex uint64) {

	for _, notificationCh := range journal.notifyOfAllCommitChangeSubscribers {

		notificationCh <- commitIndex
	}
}

func (journal *ArrayJournal) notifyOneTimeSubscribers(commitIndex uint64) {

	for subscribedIndex, notificationChs := range journal.oneTimeCommitChangeSubscribers {

		if commitIndex >= subscribedIndex {

			journal.notifyAllRegisteredSubscribersForIndex(subscribedIndex, notificationChs)
		}
	}
}

func (journal *ArrayJournal) notifyAllRegisteredSubscribersForIndex(subscribedIndex uint64,
	notificationChs []chan bool) {

	// Clean these up before notifying subscribers
	delete(journal.oneTimeCommitChangeSubscribers, subscribedIndex)

	for _, notificationCh := range notificationChs {

		notificationCh <- true
	}
}

func (journal *ArrayJournal) isCommitIndexWithinValidRange(index uint64) bool {

	return index <= uint64(journal.headIndex)
}

func (journal *ArrayJournal) isEmptyLog() bool {

	return journal.headIndex == -1
}

func (journal *ArrayJournal) validateIndexes(beginIndex uint64, endIndex uint64) error {

	var err error

	logLen := uint64(len(journal.theLog))

	if beginIndex > endIndex {

		// TODO logging here
		err = ErrInvertedIndexes
	} else if !indexesInBounds(beginIndex, endIndex, logLen) {
		// TODO logging here
		err = ErrIndexOutOfBounds
	}

	return err
}

//nolint:gocyclo
func (journal *ArrayJournal) processCommands() {

	for {

		command := <-journal.workQueue

		switch command.name {

		case Append:

			journal.append(command)
		case Commit:

			journal.commit(command)

		case GetAllUncommittedEntries:

			journal.getAllUncommittedEntries(command)
		}
	}
}

func (journal *ArrayJournal) getAllUncommittedEntries(command ArrayJournalCommand) {

	var uncommittedEntries *ArrayJournalIterator

	if journal.hasUncommittedEntries() {

		var backingArray []Entry

		if journal.isUncommitted() {

			backingArray = journal.theLog[0:]
		} else {
			backingArray = journal.theLog[journal.commitIndex+1 : journal.headIndex+1]
		}

		uncommittedEntries = NewArrayJournalIterator(backingArray)
	} else {
		uncommittedEntries = NewArrayJournalIterator([]Entry{})
	}

	result := AllUncommittedEntriesResult{
		CommitIndex:           int(journal.commitIndex),
		HasUncommittedEntries: journal.headIndex != journal.commitIndex,
		HeadIndex:             int(journal.headIndex),
		UncommittedEntries:    uncommittedEntries,
	}

	command.getAllUncommittedEntriesDoneCh <- result
}

func (journal *ArrayJournal) commit(command ArrayJournalCommand) {

	var err error

	if journal.isEmptyLog() {
		err = ErrCommitOnEmptyLog
	} else if journal.isCommitIndexWithinValidRange(uint64(command.commitIndex)) {

		journal.commitIndex = command.commitIndex
		journal.notifySubscribers(uint64(command.commitIndex))
	} else {
		err = ErrIndexBeyondHead
	}

	result := CommitResult{
		Error: err,
	}

	command.commitDoneCh <- result
}

func (journal *ArrayJournal) append(command ArrayJournalCommand) {

	journal.theLog = append(journal.theLog, command.appendEntry)
	journal.headIndex += 1

	result := AppendResult{
		Error: nil,
		Index: uint64(journal.headIndex),
	}

	command.appendDoneCh <- result
}

func indexesInBounds(beginIndex uint64, endIndex uint64, logLen uint64) bool {

	return beginIndex < logLen && endIndex <= logLen
}

func (journal *ArrayJournal) isUncommitted() bool {

	return journal.commitIndex == -1
}

func (journal *ArrayJournal) hasUncommittedEntries() bool {

	return journal.headIndex >= 0 && journal.headIndex != journal.commitIndex
}
