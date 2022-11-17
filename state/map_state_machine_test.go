package state

import (
	"github.com/jrobison153/raft/crypto"
	"github.com/jrobison153/raft/journal"
	"reflect"
	"testing"
	"time"
)

func TestWhenMapStateMachineIsCreatedThenTheHighestSeenCommitIndexIsInitialized(t *testing.T) {

	_, stateMachine := setup()

	if stateMachine.highestSeenCommitIndex != -1 {
		t.Errorf("Highest seen commit index should have defaulted to -1 but was %d",
			stateMachine.highestSeenCommitIndex)
	}
}

func TestWhenTheStateMachineIsStartedThenItRegistersForAllCommitChangesOnTheJournal(t *testing.T) {

	journalSpy, stateMachine := setup()

	expectedNotifyCh := stateMachine.commitListenCh

	if !journalSpy.IsChannelRegisteredForAllCommitChanges(expectedNotifyCh) {
		t.Errorf("State machine did not register for all commits on the Journal")
	}
}

func TestWhenStateMachineAlreadyStartedThenSubsequentCallsToStartReturnAnError(t *testing.T) {

	_, stateMachine := setup()

	err := stateMachine.Start()

	if ErrAlreadyRunning != err {
		t.Errorf("State machine already running, should have seen error '%v' but got '%v'",
			ErrAlreadyRunning,
			err)
	}
}

func TestWhenStateMachineAlreadyStartedThenSubsequentCallsToStartDoNotRegisterAgain(t *testing.T) {

	journalSpy, stateMachine := setup()

	stateMachine.Start()

	if journalSpy.NotifyOfAllCommitChangesCallCount() != 1 {
		t.Errorf("State machine should have registered for changes once but regiser count was %d",
			journalSpy.NotifyOfAllCommitChangesCallCount())
	}
}

func TestWhenStateMachineIsStartedThenItInitializesStateFromJournal(t *testing.T) {

	journalSpy := journal.NewJournalSpy()
	stateMachine := NewMapStateMachine(journalSpy)

	initData := map[string][]byte{
		"a": []byte("a value"),
		"b": []byte("b value"),
		"c": []byte("c value"),
	}

	var commitIndex uint64
	for k, v := range initData {

		hashedKey := crypto.HashIt(k)
		result := journalSpy.Append(k, hashedKey, v)

		appendResult := <-result

		commitIndex = appendResult.Index
	}

	_ = journalSpy.Commit(commitIndex)

	_ = stateMachine.Start()

	for k, v := range initData {

		if !reflect.DeepEqual(stateMachine.data[k], v) {
			t.Errorf("State machine should have been initialized with key '%s' and value '%s'", k, v)
		}
	}
}

func TestWhenMultipleCommitsThenTheStateMachineRepresentsTheResultOfTheLastCommit(t *testing.T) {

	journalSpy, stateMachine := setup()

	key := "a key"
	hashedKey := crypto.HashIt(key)

	var appendDoneCh chan journal.AppendResult
	var appendResult journal.AppendResult

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("a"))
	<-appendDoneCh

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("b"))
	appendResult = <-appendDoneCh

	midIndex := appendResult.Index

	var commitDoneCh chan journal.CommitResult
	commitDoneCh = journalSpy.Commit(midIndex)
	<-commitDoneCh

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("c"))
	<-appendDoneCh

	expectedValue := []byte("i am some data")
	appendDoneCh = journalSpy.Append(key, hashedKey, expectedValue)
	appendResult = <-appendDoneCh

	commitDoneCh = journalSpy.Commit(appendResult.Index)
	<-commitDoneCh

	// gross but need to give listener time to update state before reading, eventual consistency :)
	time.Sleep(50 * time.Millisecond)

	actualValue, _ := stateMachine.GetValueForKey(key)

	if !reflect.DeepEqual(expectedValue, actualValue) {
		t.Errorf("State machine should have updated key '%s' value to '%s' but instead value was '%s'",
			key,
			expectedValue,
			actualValue)
	}
}

func TestWhenGettingValueForValidKeyThenErrorIsNil(t *testing.T) {

	journalSpy, stateMachine := setup()

	key := "a key"
	hashedKey := crypto.HashIt(key)
	expectedValue := []byte("i am some data")

	appendDoneCh := journalSpy.Append(key, hashedKey, expectedValue)
	appendResult := <-appendDoneCh

	commitDoneCh := journalSpy.Commit(appendResult.Index)
	<-commitDoneCh

	// gross but need to give listener time to update state before reading, eventual consistency :)
	time.Sleep(50 * time.Millisecond)

	_, err := stateMachine.GetValueForKey(key)

	if err != nil {
		t.Errorf("Error should have been nil but was '%v'", err)
	}
}

func TestWhenStateMachineUpdatesDueToCommitIndexChangeThenHighestSeenCommitIndexIsUpdated(t *testing.T) {

	journalSpy, stateMachine := setup()

	key := "a key"
	hashedKey := crypto.HashIt(key)

	var appendDoneCh chan journal.AppendResult

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("a"))
	<-appendDoneCh

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("b"))
	<-appendDoneCh

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("c"))
	<-appendDoneCh

	appendDoneCh = journalSpy.Append(key, hashedKey, []byte("d"))
	appendResult := <-appendDoneCh

	commitDoneCh := journalSpy.Commit(appendResult.Index)
	<-commitDoneCh

	// gross but need to give listener time to update state before reading, eventual consistency :)
	time.Sleep(50 * time.Millisecond)

	if uint64(stateMachine.highestSeenCommitIndex) != appendResult.Index {
		t.Errorf("State machine should have updated its highest seen commit index to value %d but it was %d",
			appendResult.Index,
			stateMachine.highestSeenCommitIndex)
	}
}

func TestWhenGettingValueAndKeyDoesNotExistThenAnErrorIsReturned(t *testing.T) {

	_, stateMachine := setup()

	_, err := stateMachine.GetValueForKey("does not exist")

	if ErrKeyNotFound != err {
		t.Errorf("'Key does not exist, expected error '%v' but got '%v'",
			ErrKeyNotFound,
			err)
	}
}

func TestWhenGettingValueAndKeyDoesNotExistThenAnNilValueIsReturned(t *testing.T) {

	_, stateMachine := setup()

	val, _ := stateMachine.GetValueForKey("does not exist")

	if val != nil {
		t.Errorf("'Key does not exist, expected nil value but got '%v'", val)
	}
}

func setup() (*journal.Spy, *MapStateMachine) {

	journalSpy := journal.NewJournalSpy()

	stateMachine := NewMapStateMachine(journalSpy)

	stateMachine.Start()

	return journalSpy, stateMachine
}
