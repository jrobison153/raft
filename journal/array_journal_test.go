package journal

import (
	"bytes"
	"errors"
	"testing"
)

func TestWhenConstructedThenTheAppendQueueIsInitialized(t *testing.T) {

	logger := NewArrayJournal()

	queueCapacity := cap(logger.workQueue)

	if queueCapacity <= 0 {
		t.Errorf("Append queue should have capacity greater than zero but was %d", queueCapacity)
	}
}

func TestWhenDataAppendedToTheLogThenTheHeadEntryIsSet(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(1)

	headEntry, _ := testContext.journal.GetHead()

	if !bytes.Equal(testContext.item, headEntry.Item) {
		t.Errorf("The last appended item of the log should have been '%s' but got '%s'",
			string(testContext.item),
			string(headEntry.Item))
	}
}

func TestGivenLogEmptyWhenGettingHeadThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	_, appendErr := testContext.journal.GetHead()

	if !errors.Is(appendErr, ErrEmptyLog) {
		t.Errorf("Append should have returned an ErrEmptyLog, instead we got %v", appendErr)
	}
}

func TestWhenItemAppendedThenTheIndexIsReturned(t *testing.T) {

	testContext := setup()

	index := testContext.appendEntries(3)

	if index != 2 {
		t.Errorf("Index should have been 2 but was %d", index)
	}
}

func TestWhenIndexIsCommittedThenTheNewCommitIndexIsSet(t *testing.T) {

	testContext := setup()

	index := uint64(1)

	testContext.appendEntries(2)

	result := testContext.journal.Commit(index)

	<-result

	if index != uint64(testContext.journal.commitIndex) {
		t.Errorf("Commit index should have been set to %d but was %d", index, testContext.journal.commitIndex)
	}
}

func TestWhenAttemptingToSubscribeToCommitIndexThatIsBeyondTheHeadOfTheLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(3)

	err := testContext.journal.NotifyOfCommitOnIndexOnce(99, make(chan bool))

	if ErrIndexBeyondHead != err {
		t.Errorf("Should have received error '%v' but instead got '%v'",
			ErrIndexBeyondHead,
			err)
	}
}

func TestWhenAttemptingToSubscribeToCommitIndexOnEmptyLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	err := testContext.journal.NotifyOfCommitOnIndexOnce(4, make(chan bool))

	if ErrSubscriptionOnEmptyLog != err {
		t.Errorf("Should have received error '%v' but instead got '%v'",
			ErrIndexBeyondHead,
			err)
	}
}

func TestGivenSubscriptionForCommitIndexWhenIndexIsCommittedThenChannelIsNotified(t *testing.T) {

	testContext := setup()

	subNotifyCh := make(chan bool)

	testContext.appendEntries(2)

	_ = testContext.journal.NotifyOfCommitOnIndexOnce(1, subNotifyCh)

	go testContext.journal.Commit(1)

	subscriberNotified := <-subNotifyCh

	if !subscriberNotified {

		t.Errorf("Subscriber should have been notified of commit")
	}
}

func TestGivenSubscriptionForCommitIndexWhenFutureIndexIsCommittedThenChannelIsNotified(t *testing.T) {

	testContext := setup()

	subNotifyCh := make(chan bool)

	testContext.appendEntries(3)

	_ = testContext.journal.NotifyOfCommitOnIndexOnce(1, subNotifyCh)

	go testContext.journal.Commit(2)

	commitDone := <-subNotifyCh

	if !commitDone {

		t.Errorf("Commit should have happened successfully")
	}
}

func TestGivenMultipleSubscriptionsForDifferentCommitIndexWhenIndexIsCommittedThenChannelsAreNotified(t *testing.T) {

	testContext := setup()

	subscriberOneNotifyCh, subscriberTwoNotifyCh := testContext.setupMultipleOneTimeNotifications(4, 9, 14)

	var commitDoneOne bool
	var commitDoneTwo bool

	for i := 0; i < 2; i++ {

		select {
		case commitDoneOne = <-subscriberOneNotifyCh:
		case commitDoneTwo = <-subscriberTwoNotifyCh:
		}
	}

	if !commitDoneOne || !commitDoneTwo {
		t.Errorf("Commit should have happened successfully")
	}
}

func TestWhenIndexIsCommittedThenNotifiedSubscribersAreRemoved(t *testing.T) {

	testContext := setup()

	subscriberOneNotifyCh, subscriberTwoNotifyCh := testContext.setupMultipleOneTimeNotifications(4, 9, 14)

	var subscriberFourGone bool
	var subscriberNineGone bool

	for i := 0; i < 2; i++ {

		select {
		case <-subscriberOneNotifyCh:
			subscriberFourGone = testContext.journal.oneTimeCommitChangeSubscribers[4] == nil
		case <-subscriberTwoNotifyCh:
			subscriberNineGone = testContext.journal.oneTimeCommitChangeSubscribers[9] == nil
		}
	}

	if !subscriberFourGone || !subscriberNineGone {
		t.Errorf("subsribers should have been removed but they were not")
	}
}

func TestWhenMultipleOneTimeNotificationSubscribersForSameIndexThenAllSubscribersAreNotifiedOnCommit(t *testing.T) {

	testContext := setup()

	subscriberOneNotifyCh, subscriberTwoNotifyCh :=
		testContext.setupMultipleOneTimeNotifications(3, 3, 3)

	var commitDoneOne bool
	var commitDoneTwo bool

	for i := 0; i < 2; i++ {

		select {
		case commitDoneOne = <-subscriberOneNotifyCh:
		case commitDoneTwo = <-subscriberTwoNotifyCh:
		}
	}

	if !commitDoneOne || !commitDoneTwo {
		t.Errorf("Commit should have happened successfully")
	}
}

func TestWhenCommitIndexIsBeyondTheHeadOfEntryOfTheLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(2)

	err := testContext.journal.Commit(10)

	if err == nil {
		t.Errorf("Expected error %v but didn't get any errors", ErrIndexBeyondHead)
	}
}

func TestWhenAttemptingToCommitIndexOnAnEmptyLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	err := testContext.journal.Commit(10)

	if err == nil {
		t.Errorf("Expected error %v but didn't get any errors", ErrCommitOnEmptyLog)
	}
}

func TestWhenGettingAllCommittedEntriesThenAllCommittedEntriesAreReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(10)

	result := testContext.journal.Commit(4)

	<-result

	committedEntriesIt := testContext.journal.GetAllCommittedEntries()
	numCommittedEntries := committedEntriesIt.Size()

	if numCommittedEntries != 5 {
		t.Errorf("Should have retrieved committed entries list of size %d but was %d",
			5,
			numCommittedEntries)
	}
}

func TestWhenGettingEntriesBetweenJournalIndexesThenTheCorrectNumberOfEntriesAreReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(10)

	journalEntries, _ := testContext.journal.GetAllEntriesBetween(4, 6)

	numEntries := journalEntries.Size()

	if 3 != numEntries {
		t.Errorf("Should have retrieved %d entries but instead got %d", 3, numEntries)
	}
}

func TestWhenGettingEntriesBetweenJournalIndexesAndBeginIndexGreaterThanEndIndexThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(10)

	_, err := testContext.journal.GetAllEntriesBetween(6, 4)

	if ErrInvertedIndexes != err {
		t.Errorf("Should have received error '%s' but instead got '%s'",
			ErrInvertedIndexes,
			err)
	}
}

func TestWhenGettingEntriesBetweenJournalIndexesAndBeginIndexEqualToLogLengthThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(5)

	_, err := testContext.journal.GetAllEntriesBetween(5, 13)

	if ErrIndexOutOfBounds != err {
		t.Errorf("Should have received error '%s' but instead got '%s'",
			ErrIndexOutOfBounds,
			err)
	}
}

func TestWhenGettingEntriesBetweenJournalIndexesAndBeginIndexGreaterThanLogLengthThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(5)

	_, err := testContext.journal.GetAllEntriesBetween(6, 13)

	if ErrIndexOutOfBounds != err {
		t.Errorf("Should have received error '%s' but instead got '%s'",
			ErrIndexOutOfBounds,
			err)
	}
}

func TestWhenGettingEntriesBetweenJournalIndexesAndEndIndexGreaterThanLogLengthThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(10)

	_, err := testContext.journal.GetAllEntriesBetween(1, 13)

	if ErrIndexOutOfBounds != err {
		t.Errorf("Should have received error '%s' but instead got '%s'",
			ErrIndexOutOfBounds,
			err)
	}
}

func TestGivenSubscriptionForAllCommitsWhenIndexCommittedThenAllSubscribersAreNotified(t *testing.T) {

	testContext := setup()

	testContext.appendEntries(10)

	subscriberOneCh := make(chan uint64)
	testContext.journal.NotifyOfAllCommitChanges(subscriberOneCh)

	subscriberTwoCh := make(chan uint64)
	testContext.journal.NotifyOfAllCommitChanges(subscriberTwoCh)

	testContext.journal.Commit(3)

	var subscriberOneIndex uint64
	var subscriberTwoIndex uint64

	for i := 0; i < 2; i++ {

		select {
		case subscriberOneIndex = <-subscriberOneCh:
		case subscriberTwoIndex = <-subscriberTwoCh:
		}
	}

	if subscriberOneIndex != 3 || subscriberTwoIndex != 3 {
		t.Errorf("Both one time commit change subscribers should have received the same commit index 3"+
			" but subscriber one got %d and subscirber two got %d",
			subscriberOneIndex,
			subscriberTwoIndex)
	}
}

func TestWhenNoUncommittedEntriesThenGetAllUncommittedEntriesResultHasUncommittedEntriesIsFalse(t *testing.T) {

	_, uncommittedEntriesResult := setupGetUncommittedEntries(0, false)

	if uncommittedEntriesResult.HasUncommittedEntries {
		t.Errorf("There weren't any uncommitted entries, expected value to be False")
	}
}

func TestWhenUncommittedEntriesThenGetAllUncommittedEntriesResultHasUncommittedEntriesIsTrue(t *testing.T) {

	_, uncommittedEntriesResult := setupGetUncommittedEntries(5, false)

	if !uncommittedEntriesResult.HasUncommittedEntries {
		t.Errorf("There were uncommitted entries, expected value to be True")
	}
}

func TestWhenGettingUncommittedEntriesThenCurrentCommitIndexIsReturned(t *testing.T) {

	numEntriesToAppend := 5
	expectedCommitIndex := numEntriesToAppend - 1

	_, uncommittedEntriesResult := setupGetUncommittedEntries(numEntriesToAppend, true)

	if expectedCommitIndex != uncommittedEntriesResult.CommitIndex {

		t.Errorf("Current commit index should have been %d but was %d",
			expectedCommitIndex,
			uncommittedEntriesResult.CommitIndex)
	}
}

func TestWhenGettingUncommittedEntriesThenCurrentHeadIndexIsReturned(t *testing.T) {

	numEntriesToAppend := 5
	expectedHeadIndex := numEntriesToAppend - 1

	_, uncommittedEntriesResult := setupGetUncommittedEntries(numEntriesToAppend, false)

	if expectedHeadIndex != uncommittedEntriesResult.HeadIndex {

		t.Errorf("Head index should have been %d but was %d",
			expectedHeadIndex,
			uncommittedEntriesResult.HeadIndex)
	}
}

func TestWhenEmptyJournalAndThereAreNoEntriesToCommitThenGetAllUncommittedEntriesIteratorHasNothingToIterateOver(t *testing.T) {

	_, uncommittedEntriesResult := setupGetUncommittedEntries(0, false)

	if uncommittedEntriesResult.UncommittedEntries.Size() != 0 {
		t.Errorf("There should have been no uncommitted entries, but we got at least one")
	}
}

func TestWhenNewJournalAndThereAreEntriesToCommitThenGetAllUncommittedEntriesIteratorIncludesThemAll(t *testing.T) {

	numEntriesToAppend := 5
	_, uncommittedEntriesResult := setupGetUncommittedEntries(numEntriesToAppend, false)

	if numEntriesToAppend != uncommittedEntriesResult.UncommittedEntries.Size() {
		t.Errorf("There should have been %d uncommitted entries, but we got %d",
			numEntriesToAppend,
			uncommittedEntriesResult.UncommittedEntries.Size())
	}
}

func TestWhenNonEmptyJournalAndThereAreNoEntriesToCommitThenGetAllUncommittedEntriesIteratorHasNothingToIterateOver(t *testing.T) {

	_, uncommittedEntriesResult := setupGetUncommittedEntries(5, true)

	if uncommittedEntriesResult.UncommittedEntries.Size() != 0 {
		t.Errorf("There should have been no uncommitted entries, but we got at least one")
	}
}

func TestWhenNonEmptyJournalAndThereAreEntriesToCommitThenGetAllUncommittedEntriesIteratorIncludesThemAll(t *testing.T) {

	testContext, _ := setupGetUncommittedEntries(5, true)

	uncommittedAppendEntries := 4
	testContext.appendEntries(4)

	uncommittedEntriesResultCh := testContext.journal.GetAllUncommittedEntries()
	uncommittedEntriesResult := <-uncommittedEntriesResultCh

	if uncommittedAppendEntries != uncommittedEntriesResult.UncommittedEntries.Size() {
		t.Errorf("There should have been %d uncommitted entries, but we got %d",
			uncommittedAppendEntries,
			uncommittedEntriesResult.UncommittedEntries.Size())
	}
}

type TestContext struct {
	item     []byte
	journal  *ArrayJournal
	mutexSpy *MutexSpy
}

func setup() *TestContext {

	item := []byte("I am some sexy Item!")

	mutexSpy := NewMutexSpy()

	logger := NewArrayJournal()

	testContext := &TestContext{
		item:     item,
		journal:  logger,
		mutexSpy: mutexSpy,
	}

	return testContext
}

func setupGetUncommittedEntries(numEntriesToAppend int, commitAfterAppend bool) (*TestContext, AllUncommittedEntriesResult) {

	testContext := setup()

	index := testContext.appendEntries(numEntriesToAppend)

	if commitAfterAppend {

		commitDoneCh := testContext.journal.Commit(index)
		<-commitDoneCh
	}

	uncommittedEntriesResultCh := testContext.journal.GetAllUncommittedEntries()
	uncommittedEntriesResult := <-uncommittedEntriesResultCh

	return testContext, uncommittedEntriesResult
}

func (testContext *TestContext) setupMultipleOneTimeNotifications(notifyIndexOne uint64,
	notifyIndexTwo uint64,
	commitIndex uint64) (chan bool, chan bool) {

	subscriberOneNotifyCh := make(chan bool)
	subscriberTwoNotifyCh := make(chan bool)

	testContext.appendEntries(15)

	testContext.journal.NotifyOfCommitOnIndexOnce(notifyIndexOne, subscriberOneNotifyCh)
	testContext.journal.NotifyOfCommitOnIndexOnce(notifyIndexTwo, subscriberTwoNotifyCh)

	testContext.journal.Commit(commitIndex)

	return subscriberOneNotifyCh, subscriberTwoNotifyCh
}

func (testContext *TestContext) appendEntries(count int) uint64 {

	var index uint64

	for i := 0; i < count; i++ {
		appendResultCh := testContext.journal.Append(testContext.item)

		result := <-appendResultCh

		index = result.Index
	}

	return index
}
