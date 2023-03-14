package state

import (
	"encoding/json"
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

	rawKeyValData := map[string][]byte{
		"a": []byte("a value"),
		"b": []byte("b value"),
		"c": []byte("c value"),
	}

	loadJournalAndCommit(rawKeyValData, journalSpy)

	_ = stateMachine.Start()

	for k, v := range rawKeyValData {

		if !reflect.DeepEqual(stateMachine.data[k], v) {
			t.Errorf("State machine should have been initialized with key '%s' and value '%s'", k, v)
		}
	}
}

func TestWhenStateMachineIsStartedAndKeyValUnmarshalFailsThenAnErrorIsReturned(t *testing.T) {

	_, err := setupForBadJournalData()

	if ErrCannotUnmarshalKeyValData != err {
		t.Errorf("Should have received error %v but got %v", ErrCannotUnmarshalKeyValData, err)
	}
}

func TestWhenStateMachineIsStartedAndKeyValUnmarshalFailsThenStateMachineIsNotStarted(t *testing.T) {

	stateMachine, _ := setupForBadJournalData()

	if stateMachine.isRunning {
		t.Errorf("State machine should not have started due to error")
	}
}

func TestWhenMultipleCommitsThenTheStateMachineRepresentsTheResultOfTheLastCommit(t *testing.T) {

	journalSpy := journal.NewJournalSpy()
	stateMachine := NewMapStateMachine(journalSpy)

	rawKeyValData := map[string][]byte{
		"a": []byte("a value"),
		"b": []byte("b value"),
		"c": []byte("c value"),
	}

	loadJournalAndCommit(rawKeyValData, journalSpy)

	expectedData := []byte("b value has been updated")
	rawKeyValDataUpdate := map[string][]byte{
		"b": expectedData,
		"c": []byte("c value"),
	}

	loadJournalAndCommit(rawKeyValDataUpdate, journalSpy)

	_ = stateMachine.Start()

	key := Key{
		Key: "b",
	}

	marshalledKey, _ := json.Marshal(key)

	actualData, _ := stateMachine.ResolveRequestToData(marshalledKey)

	if !reflect.DeepEqual(expectedData, actualData) {
		t.Errorf("State machine should have updated key 'b' value to '%s' but instead value was '%s'",
			expectedData,
			actualData)
	}
}

func TestWhenGettingValueForValidKeyThenErrorIsNil(t *testing.T) {

	journalSpy, stateMachine := setup()

	rawKeyValData := map[string][]byte{
		"a": []byte("a value"),
		"b": []byte("b value"),
		"c": []byte("c value"),
	}

	loadJournalAndCommit(rawKeyValData, journalSpy)

	// gross but need to give listener time to update state before reading, eventual consistency :)
	time.Sleep(50 * time.Millisecond)

	marshalledKey := createKeyRequest("b")

	_, err := stateMachine.ResolveRequestToData(marshalledKey)

	if err != nil {
		t.Errorf("Error should have been nil but was '%v'", err)
	}
}

func TestWhenStateMachineUpdatesDueToCommitIndexChangeThenHighestSeenCommitIndexIsUpdated(t *testing.T) {

	journalSpy, stateMachine := setup()

	rawKeyValData := map[string][]byte{
		"a": []byte("a value"),
		"b": []byte("b value"),
		"c": []byte("c value"),
	}

	expectedCommitIndex := loadJournalAndCommit(rawKeyValData, journalSpy)

	// gross but need to give listener time to update state before reading, eventual consistency :)
	time.Sleep(50 * time.Millisecond)

	if uint64(stateMachine.highestSeenCommitIndex) != expectedCommitIndex {
		t.Errorf("State machine should have updated its highest seen commit index to value %d but it was %d",
			expectedCommitIndex,
			stateMachine.highestSeenCommitIndex)
	}
}

func TestWhenGettingValueAndKeyDoesNotExistThenAnErrorIsReturned(t *testing.T) {

	_, stateMachine := setup()

	requestItem := createKeyRequest("does not exist key")

	_, err := stateMachine.ResolveRequestToData(requestItem)

	if ErrKeyNotFound != err {
		t.Errorf("'Key does not exist, expected error '%v' but got '%v'",
			ErrKeyNotFound,
			err)
	}
}

func TestWhenGettingValueAndKeyDoesNotExistThenNilValueIsReturned(t *testing.T) {

	_, stateMachine := setup()

	requestItem := createKeyRequest("does not exist key")

	val, _ := stateMachine.ResolveRequestToData(requestItem)

	if val != nil {
		t.Errorf("'Key does not exist, expected nil value but got '%v'", val)
	}
}
func TestWhenResolvingRequestToGetDataAndRequestIsNotCorrectStructureThenAnErrorIsReturned(t *testing.T) {

	_, stateMachine := setup()

	requestItem := []byte("not valid request")

	_, err := stateMachine.ResolveRequestToData(requestItem)

	if ErrInvalidKeyRequest != err {
		t.Errorf("Expected error '%v' but got '%v'", ErrInvalidKeyRequest, err)
	}
}

func setup() (*journal.Spy, *MapStateMachine) {

	journalSpy := journal.NewJournalSpy()

	stateMachine := NewMapStateMachine(journalSpy)

	stateMachine.Start()

	return journalSpy, stateMachine
}

func loadJournalAndCommit(rawKeyValData map[string][]byte, journalSpy *journal.Spy) uint64 {

	items := make([][]byte, 0, len(rawKeyValData))

	for k, v := range rawKeyValData {

		marshalledRawData := createKeyValRequestItem(k, v)

		items = append(items, marshalledRawData)
	}

	var commitIndex uint64
	for _, v := range items {

		result := journalSpy.Append(v)

		appendResult := <-result

		commitIndex = appendResult.Index
	}

	commitDoneCh := journalSpy.Commit(commitIndex)

	<-commitDoneCh

	return commitIndex
}

func createKeyValRequestItem(k string, v []byte) []byte {

	rawData := KeyValItem{
		Key:  k,
		Data: v,
	}

	marshalledRawData, _ := json.Marshal(rawData)
	return marshalledRawData
}

func createKeyRequest(key string) []byte {

	keyRequest := Key{
		Key: key,
	}

	marshalledKey, _ := json.Marshal(keyRequest)

	return marshalledKey
}

func setupForBadJournalData() (*MapStateMachine, error) {

	journalSpy := journal.NewJournalSpy()
	stateMachine := NewMapStateMachine(journalSpy)

	notKeyValueData := []byte("certainly not key value format")

	appendDoneCh := journalSpy.Append(notKeyValueData)
	appendResult := <-appendDoneCh

	commitDoneCh := journalSpy.Commit(appendResult.Index)
	<-commitDoneCh

	err := stateMachine.Start()

	return stateMachine, err
}
