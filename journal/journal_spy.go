package journal

import (
	"errors"
)

// TODO look at refactoring this spy as it has become a fake and in reality it could wrap ArrayJournal

const (
	SpyAppend                   = "append"
	SpyCommit                   = "commit"
	SpyGetAllUncommittedEntries = "get-all-uncommitted-entries"
)

type SpyEntry struct {
	Item []byte
}

type Spy struct {
	appendCalled                           bool
	allChangesNotifyCh                     chan uint64
	appendFailMsg                          string
	commitCalled                           bool
	committedEntryIndex                    int
	isFailingNextAppend                    bool
	isFailingNextNotifyOfCommitOnIndexOnce bool
	log                                    []SpyEntry
	notifyOfAllCommitChangesCallCount      int
	spiedAppendItem                        []byte
	spiedAppendKey                         []byte
	subscribers                            map[uint64]chan bool
	workQueue                              chan SpyCommand
}

func NewJournalSpy() *Spy {

	spy := &Spy{
		log:                                    make([]SpyEntry, 0, 16),
		appendCalled:                           false,
		commitCalled:                           false,
		committedEntryIndex:                    -1,
		isFailingNextAppend:                    false,
		isFailingNextNotifyOfCommitOnIndexOnce: false,
		subscribers:                            make(map[uint64]chan bool),
		notifyOfAllCommitChangesCallCount:      0,
		workQueue:                              make(chan SpyCommand, 32),
	}

	go spy.processCommands()

	return spy
}

// Begin Journaler interface

type SpyCommand struct {
	appendEntry                    SpyEntry
	appendDoneCh                   chan AppendResult
	name                           string
	commitDoneCh                   chan CommitResult
	commitIndex                    uint64
	getAllUncommittedEntriesDoneCh chan AllUncommittedEntriesResult
}

func (spy *Spy) Append(item []byte) chan AppendResult {

	entry := SpyEntry{
		Item: item,
	}

	doneCh := make(chan AppendResult)

	command := SpyCommand{
		name:         SpyAppend,
		appendEntry:  entry,
		appendDoneCh: doneCh,
	}

	spy.workQueue <- command

	return doneCh
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

func (spy *Spy) Commit(index uint64) chan CommitResult {

	spy.commitCalled = true

	doneCh := make(chan CommitResult)
	command := SpyCommand{
		name:         SpyCommit,
		commitDoneCh: doneCh,
		commitIndex:  index,
	}

	spy.workQueue <- command

	return doneCh
}

func (spy *Spy) NotifyOfAllCommitChanges(ch chan uint64) {

	spy.allChangesNotifyCh = ch
	spy.notifyOfAllCommitChangesCallCount += 1
}

func (spy *Spy) GetHead() (Entry, error) {
	//TODO implement me
	panic("implement me")
}

// GetAllEntriesBetween startIndex and sizeOfBackingArray are inclusive
func (spy *Spy) GetAllEntriesBetween(beginIndex uint64, endIndex uint64) (Iterator, error) {

	indexToEndOfLog := spy.log[beginIndex : endIndex+1]

	resultEntries := createEntryArray(indexToEndOfLog)

	iterator := NewArrayJournalIterator(resultEntries)

	return iterator, nil
}

func (spy *Spy) GetAllCommittedEntries() Iterator {

	resultEntries := createEntryArray(spy.log)

	iterator := NewArrayJournalIterator(resultEntries)

	return iterator
}

func (spy *Spy) GetAllUncommittedEntries() chan AllUncommittedEntriesResult {

	resultCh := make(chan AllUncommittedEntriesResult)

	command := SpyCommand{
		name:                           SpyGetAllUncommittedEntries,
		getAllUncommittedEntriesDoneCh: resultCh,
	}

	spy.workQueue <- command

	return resultCh
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

	return spy.spiedAppendItem
}

func (spy *Spy) LatestEntryForKey(key []byte) SpyEntry {

	//var mostRecentAppendForKey SpyEntry

	//for _, entry := range spy.log {
	//
	//	if reflect.DeepEqual(entry.Key, key) {
	//		mostRecentAppendForKey = entry
	//		break
	//	}
	//}

	return SpyEntry{}
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

func (spy *Spy) IsChannelRegisteredForAllCommitChanges(ch chan uint64) bool {

	return spy.allChangesNotifyCh == ch
}

func (spy *Spy) NotifyOfAllCommitChangesCallCount() int {

	return spy.notifyOfAllCommitChangesCallCount
}

//nolint:gocyclo
func (spy *Spy) processCommands() {

	for {
		command := <-spy.workQueue

		switch command.name {

		case SpyGetAllUncommittedEntries:
			spy.getAllUncommittedEntries(command)
		case SpyAppend:

			spy.append(command)
		case SpyCommit:

			spy.commit(command)
		}
	}
}

func (spy *Spy) commit(command SpyCommand) {

	spy.committedEntryIndex = int(command.commitIndex)

	notifySubscribersOfIndexChange(command.commitIndex, spy.subscribers)

	if spy.allChangesNotifyCh != nil {

		spy.allChangesNotifyCh <- command.commitIndex
	}

	result := CommitResult{
		Error: nil,
	}

	command.commitDoneCh <- result
}

func (spy *Spy) append(command SpyCommand) {

	var err error

	if spy.isFailingNextAppend {
		err = errors.New(spy.appendFailMsg)
	} else {

		spy.appendCalled = true
		spy.spiedAppendItem = command.appendEntry.Item

		entry := command.appendEntry

		spy.log = append(spy.log, entry)

		err = nil
	}

	headIndex := len(spy.log) - 1

	result := AppendResult{
		Index: uint64(headIndex),
		Error: err,
	}

	command.appendDoneCh <- result
}

func (spy *Spy) CommitCalled() bool {

	return spy.commitCalled
}

func (spy *Spy) CommitCalledOnIndex(index uint64) bool {

	return spy.committedEntryIndex == int(index)
}

func (spy *Spy) getAllUncommittedEntries(command SpyCommand) {

	var uncommittedSpyEntries []SpyEntry

	numLogEntries := len(spy.log)

	if numLogEntries > 0 {

		uncommittedSpyEntries = spy.log[spy.committedEntryIndex+1 : numLogEntries]
	} else {
		uncommittedSpyEntries = []SpyEntry{}
	}

	entries := createEntryArray(uncommittedSpyEntries)
	entryIterator := NewArrayJournalIterator(entries)

	uncommittedEntriesResult := AllUncommittedEntriesResult{
		UncommittedEntries:    entryIterator,
		HasUncommittedEntries: len(uncommittedSpyEntries) > 0,
		CommitIndex:           spy.committedEntryIndex,
		HeadIndex:             numLogEntries - 1,
	}

	command.getAllUncommittedEntriesDoneCh <- uncommittedEntriesResult
}

func notifySubscribersOfIndexChange(currentCommitIndex uint64, subscribers map[uint64]chan bool) {

	for commitIndex, ch := range subscribers {

		if currentCommitIndex >= commitIndex {
			ch <- true
		}
	}
}

func createEntryArray(journal []SpyEntry) []Entry {
	resultEntries := make([]Entry, 0, 16)

	for _, k := range journal {

		anEntry := Entry(k)
		resultEntries = append(resultEntries, anEntry)
	}
	return resultEntries
}

// End Spy functions
