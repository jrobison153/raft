package client

import (
	"github.com/jrobison153/raft/journal"
	"github.com/jrobison153/raft/state"
	"reflect"
	"testing"
)

func TestWhenAnItemIsPutSuccessfullyThenThereAreNoErrors(t *testing.T) {

	testContext := setup()

	_, err := testContext.client.Put(testContext.item)

	if err != nil {
		t.Errorf("Expected no errors but got error '%s'", err)
	}
}

func TestWhenAnItemIsPutThenTheItemIsAppendedToTheLog(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.item)

	if testContext.journalSpy.AppendCalled() != true {
		t.Error("Put item was not appended to the journal")
	}
}

func TestWhenLogAppendFailsThenErrorIsReturned(t *testing.T) {

	testContext := setup()
	testContext.journalSpy.FailNextAppend("whatever")

	_, err := testContext.client.Put(testContext.item)

	if ErrAppendToJournalFailed != err {
		t.Errorf("Expected error '%v' but got '%v'", ErrAppendToJournalFailed, err)
	}
}

func TestWhenItemIsPutThenChannelReturnedUnblocksOnCommitSuccess(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put(testContext.item)

	go testContext.journalSpy.Commit(1)

	isPutComplete := <-doneCh

	if !isPutComplete {
		t.Error("Expected commit to complete successfully")
	}
}

func TestWhenLogAppendFailsThenTheCorrectErrorMessageIsReturned(t *testing.T) {

	testContext := setup()
	testContext.journalSpy.FailNextAppend("failing for test purposes")

	_, err := testContext.client.Put(testContext.item)

	if ErrAppendToJournalFailed != err {

		t.Errorf("Should have received error '%v' but got '%v'", ErrAppendToJournalFailed, err)
	}
}

func TestWhenLogAppendFailsThenReturnedChannelUnblocks(t *testing.T) {

	testContext := setup()

	testContext.journalSpy.FailNextAppend("Failing for test purposes")

	doneCh, _ := testContext.client.Put(testContext.item)

	result := <-doneCh

	if result {
		t.Error("Channel should have unblocked with result of false")
	}
}

func TestWhenItemIsPutThenSubscriptionSetupOnJournalIndex(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.item)

	if !testContext.journalSpy.RegisteredForNotifyOnIndex(0) {
		t.Errorf("Subscriber for journal index 0 should have been registered but was not")
	}
}

func TestWhenItemIsPutThenSubscriptionSetupOnJournalIndexAssociatedWithReturnedChannel(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put(testContext.item)

	if !testContext.journalSpy.RegisteredNotifyChannelOnIndex(0, doneCh) {
		t.Errorf("Channel %v should have been registered for notification with journal index 0", doneCh)
	}
}

func TestWhenItemIsPutAndSubscriptionForCommitOnIndexFailsThenTheChannelIsUnblocked(t *testing.T) {

	testContext := setup()

	testContext.journalSpy.FailNextNotifyOfCommitOnIndexOnce()

	doneCh, _ := testContext.client.Put(testContext.item)

	result := <-doneCh

	if result {
		t.Errorf("Channel should have been unblocked with false value because subscription on commit index failed")
	}
}

func TestWhenItemIsPutAndSubscriptionForCommitOnIndexFailsThenTheCorrectErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.journalSpy.FailNextNotifyOfCommitOnIndexOnce()

	_, err := testContext.client.Put(testContext.item)

	if ErrRegisterForNotificationOnCommit != err {

		t.Errorf("Should have received error '%v' but instead got '%v'",
			ErrRegisterForNotificationOnCommit,
			err)
	}
}

func TestWhenGettingDataForValidRequestThenTheDataIsReturned(t *testing.T) {

	testContext := setup()

	testContext.stateMachineSpy.AddStateMachineData(testContext.item)

	retrievedData, _ := testContext.client.Get(testContext.item)

	if !reflect.DeepEqual(testContext.item, retrievedData) {

		t.Errorf("Item %s should have been retrieved", testContext.item)
	}
}

func TestWhenGettingDataForValidRequestThenNoErrorIsReturned(t *testing.T) {

	testContext := setup()

	testContext.stateMachineSpy.AddStateMachineData(testContext.item)

	_, err := testContext.client.Get(testContext.item)

	if err != nil {

		t.Errorf("Error '%v' returned when there shouldn't have been an error", err)
	}
}

// ============================= Test Support ==================

type TestContext struct {
	client          *Client
	journalSpy      *journal.Spy
	stateMachineSpy *state.Spy
	item            []byte
}

func setup() *TestContext {

	journalSpy := journal.NewJournalSpy()
	stateMachineSpy := state.NewStateMachineSpy(journalSpy)

	client := New(journalSpy, stateMachineSpy)

	testContext := &TestContext{
		client:          client,
		journalSpy:      journalSpy,
		stateMachineSpy: stateMachineSpy,
		item:            []byte("i am some item"),
	}

	return testContext
}
