package journal

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestWhenConstructedTheMemSyncImplementationIsDefaultedToMutex(t *testing.T) {

	logger := NewArrayJournal()

	mutexType := reflect.TypeOf(&sync.Mutex{}).String()
	defaultSyncType := reflect.TypeOf(logger.locker).String()

	if strings.Compare(mutexType, defaultSyncType) != 0 {
		t.Errorf("Memory synchronization implementation should have been defaulted to '%v' but was '%v'",
			mutexType,
			defaultSyncType)
	}
}

func TestWhenDataAppendedToTheLogThenTheHeadEntryKeyIsSet(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 1)

	headEntry, _ := testContext.logger.GetHead()

	if !bytes.Equal(testContext.key, headEntry.Key) {
		t.Errorf("The last appended item of the log should have been key '%s' but got '%s'",
			string(testContext.key),
			string(headEntry.Key))
	}
}

func TestWhenDataAppendedToTheLogThenTheHeadEntryDataIsSet(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 1)

	headEntry, _ := testContext.logger.GetHead()

	if !bytes.Equal(testContext.data, headEntry.Data) {
		t.Errorf("The last appended item of the log should have been data '%s' but got '%s'",
			string(testContext.data),
			string(headEntry.Data))
	}
}

func TestGivenLogEmptyWhenGettingHeadThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	_, appendErr := testContext.logger.GetHead()

	if !errors.Is(appendErr, ErrEmptyLog) {
		t.Errorf("Append should have returned an ErrEmptyLog, instead we got %v", appendErr)
	}
}

func TestWhenGettingHeadEntryThenAccessToSharedDataIsLockedBeforeItIsUnlocked(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 10)

	testContext.mutexSpy.Reset()

	testContext.logger.GetHead()

	if !testContext.mutexSpy.UnlockCalledBeforeLock() {
		t.Errorf("Access to shared should have been locked before it was unlocked")
	}
}

func TestWhenItemAppendedThenTheIndexIsReturned(t *testing.T) {

	testContext := setup()

	index := appendEntries(testContext, 3)

	if index != 2 {
		t.Errorf("Index should have been 2 but was %d", index)
	}
}

func TestWhenItemAppendedThenAccessToSharedDataIsLockedBeforeItIsUnlocked(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 1)

	if !testContext.mutexSpy.UnlockCalledBeforeLock() {
		t.Errorf("Access to shared should have been locked before it was unlocked")
	}
}

func TestWhenIndexIsCommittedThenTheNewCommitIndexIsSet(t *testing.T) {

	testContext := setup()

	index := uint64(1)

	appendEntries(testContext, 2)

	testContext.logger.Commit(index)

	if index != testContext.logger.commitIndex {
		t.Errorf("Commit index should have been set to %d but was %d", index, testContext.logger.commitIndex)
	}
}

func TestWhenAttemptingToSubscribeToCommitIndexThatIsBeyondTheHeadOfTheLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 3)

	err := testContext.logger.NotifyOfCommitOnIndexOnce(99, make(chan bool))

	if ErrIndexBeyondHead != err {
		t.Errorf("Should have received error '%v' but instead got '%v'",
			ErrIndexBeyondHead,
			err)
	}
}

func TestWhenAttemptingToSubscribeToCommitIndexOnEmptyLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	err := testContext.logger.NotifyOfCommitOnIndexOnce(4, make(chan bool))

	if ErrSubscriptionOnEmptyLog != err {
		t.Errorf("Should have received error '%v' but instead got '%v'",
			ErrIndexBeyondHead,
			err)
	}
}

func TestWhenSubscribingToCommitIndexThenAccessToSharedDataIsLockedBeforeItIsUnlocked(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 10)

	testContext.mutexSpy.Reset()

	testContext.logger.NotifyOfCommitOnIndexOnce(4, make(chan bool))

	if !testContext.mutexSpy.UnlockCalledBeforeLock() {
		t.Errorf("Access to shared should have been locked before it was unlocked")
	}
}

func TestGivenSubscriptionForCommitIndexWhenIndexIsCommittedThenChannelIsNotified(t *testing.T) {

	testContext := setup()

	subNotifyCh := make(chan bool)

	appendEntries(testContext, 2)

	testContext.logger.NotifyOfCommitOnIndexOnce(1, subNotifyCh)

	go testContext.logger.Commit(1)

	subscriberNotified := <-subNotifyCh

	if !subscriberNotified {

		t.Errorf("Subscriber should have been notified of commit")
	}
}

func TestGivenSubscriptionForCommitIndexWhenFutureIndexIsCommittedThenChannelIsNotified(t *testing.T) {

	testContext := setup()

	subNotifyCh := make(chan bool)

	appendEntries(testContext, 3)

	testContext.logger.NotifyOfCommitOnIndexOnce(1, subNotifyCh)

	go testContext.logger.Commit(2)

	commitDone := <-subNotifyCh

	if !commitDone {

		t.Errorf("Commit should have happened successfully")
	}
}

func TestGivenMultipleSubscriptionsForDifferentCommitIndexWhenIndexIsCommittedThenChannelsAreNotified(t *testing.T) {

	commitDoneChOne, commitDoneChTwo := setupMultipleNotifications()

	var commitDoneOne bool
	var commitDoneTwo bool

	for i := 0; i < 2; i++ {

		select {
		case commitDoneOne = <-commitDoneChOne:
		case commitDoneTwo = <-commitDoneChTwo:
		}
	}

	if !commitDoneOne || !commitDoneTwo {
		t.Errorf("Commit should have happened successfully")
	}
}

func TestWhenIndexIsCommittedThenNotifiedSubscribersAreRemoved(t *testing.T) {

	testContext := setup()

	commitDoneChOne, commitDoneChTwo := setupMultipleNotifications()

	var subscriberFourGone bool
	var subscriberNineGone bool

	for i := 0; i < 2; i++ {

		select {
		case <-commitDoneChOne:
			subscriberFourGone = testContext.logger.subscribers[4] == nil
		case <-commitDoneChTwo:
			subscriberNineGone = testContext.logger.subscribers[9] == nil
		}
	}

	if !subscriberFourGone || !subscriberNineGone {
		t.Errorf("subsribers should have been removed but they were not")
	}
}

func TestWhenCommitIndexIsBeyondTheHeadOfEntryOfTheLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 2)

	err := testContext.logger.Commit(10)

	if err == nil {
		t.Errorf("Expected error %v but didn't get any errors", ErrIndexBeyondHead)
	}
}

func TestWhenAttemptingToCommitIndexOnAnEmptyLogThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	err := testContext.logger.Commit(10)

	if err == nil {
		t.Errorf("Expected error %v but didn't get any errors", ErrCommitOnEmptyLog)
	}
}

func TestWhenIndexIsCommittedThenAccessToSharedDataIsLockedBeforeItIsUnlocked(t *testing.T) {

	testContext := setup()

	appendEntries(testContext, 10)

	testContext.mutexSpy.Reset()

	testContext.logger.Commit(4)

	if !testContext.mutexSpy.UnlockCalledBeforeLock() {
		t.Errorf("Access to shared should have been locked before it was unlocked")
	}
}

func appendEntries(testContext *TestContext, count int) uint64 {

	var index uint64

	for i := 0; i < count; i++ {
		index, _ = testContext.logger.Append(testContext.key, testContext.data)
	}

	return index
}

type TestContext struct {
	key      []byte
	data     []byte
	logger   *ArrayJournal
	mutexSpy *MutexSpy
}

func setup() *TestContext {

	key := []byte("12345")
	data := []byte("I am some sexy data!")

	mutexSpy := NewMutexSpy()

	logger := NewArrayJournal()
	logger.SetMemoryLocker(mutexSpy)

	testContext := &TestContext{
		key:      key,
		data:     data,
		logger:   logger,
		mutexSpy: mutexSpy,
	}

	return testContext
}

func setupMultipleNotifications() (chan bool, chan bool) {

	testContext := setup()

	commitDoneChOne := make(chan bool)
	commitDoneChTwo := make(chan bool)

	appendEntries(testContext, 15)

	testContext.logger.NotifyOfCommitOnIndexOnce(4, commitDoneChOne)
	testContext.logger.NotifyOfCommitOnIndexOnce(9, commitDoneChTwo)

	go testContext.logger.Commit(14)

	return commitDoneChOne, commitDoneChTwo
}
