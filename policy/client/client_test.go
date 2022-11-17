package client

import (
	"crypto/sha1"
	"github.com/jrobison153/raft/journal"
	"reflect"
	"testing"
)

// Missing tests
// * Put - setup notify of commit
// * Put - notify of commit fail close ch
// * Put - notify of commit fail return ErrUnableToRegisterCommitIndex

func TestWhenAnItemIsPutSuccessfullyThenThereAreNoErrors(t *testing.T) {

	testContext := setup()

	_, err := testContext.client.Put(testContext.key, testContext.data)

	if err != nil {
		t.Errorf("Expected no errors but got error '%s'", err)
	}
}

func TestWhenAnItemIsPutThenTheItemIsAppendedToTheLog(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.key, testContext.data)

	if testContext.journalSpy.AppendCalled() != true {
		t.Error("Put item was not appended to the journal")
	}
}

func TestWhenItemIsPutThenTheKeyIsHashed(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.key, testContext.data)

	spiedKey := testContext.journalSpy.AppendKey()
	expectedKey := hashIt(testContext.key)

	if !reflect.DeepEqual(spiedKey, expectedKey) {
		t.Errorf(
			"Key not correctly hashed when passed to the Log. Expected '%s' got '%s'",
			expectedKey,
			spiedKey)
	}
}

func TestWhenItemIsPutThenTheDataIsAssociatedWithTheKey(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.key, testContext.data)
	hashedKey := hashIt(testContext.key)

	spiedData := testContext.journalSpy.LatestDataForKey(hashedKey)
	expectedData := testContext.data

	if !reflect.DeepEqual(spiedData, expectedData) {
		t.Errorf("Data was not correctly associated with the key. Expected '%s', got '%s' for key '%s'",
			expectedData,
			spiedData,
			testContext.key)
	}
}

func TestWhenLogAppendFailsThenErrorIsReturned(t *testing.T) {

	testContext := setup()
	testContext.journalSpy.FailNextAppend("whatever")

	_, err := testContext.client.Put(testContext.key, testContext.data)

	if ErrAppendToLogFailed != err {
		t.Errorf("Expected error '%v' but got '%v'", ErrAppendToLogFailed, err)
	}
}

func TestWhenItemIsPutThenChannelReturnedUnblocksOnCommitSuccess(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	go testContext.journalSpy.Commit(1)

	isPutComplete := <-doneCh

	if !isPutComplete {
		t.Error("Expected commit to complete successfully")
	}
}

func TestWhenLogAppendFailsThenTheCorrectErrorMessageIsReturned(t *testing.T) {

	testContext := setup()
	testContext.journalSpy.FailNextAppend("failing for test purposes")

	_, err := testContext.client.Put(testContext.key, testContext.data)

	if ErrAppendToLogFailed != err {

		t.Errorf("Should have received error '%v' but got '%v'", ErrAppendToLogFailed, err)
	}
}

func TestWhenLogAppendFailsThenReturnedChannelUnblocks(t *testing.T) {

	testContext := setup()

	testContext.journalSpy.FailNextAppend("Failing for test purposes")

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	result := <-doneCh

	if result {
		t.Error("Channel should have unblocked with result of false")
	}
}

func TestWhenItemIsPutThenSubscriptionSetupOnJournalIndex(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.key, testContext.data)

	if !testContext.journalSpy.RegisteredForNotifyOnIndex(0) {
		t.Errorf("Subscriber for journal index 0 should have been registered but was not")
	}
}

func TestWhenItemIsPutThenSubscriptionSetupOnJournalIndexAssociatedWithReturnedChannel(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	if !testContext.journalSpy.RegisteredNotifyChannelOnIndex(0, doneCh) {
		t.Errorf("Channel %v should have been registered for notification with journal index 0", doneCh)
	}
}

func TestWhenItemIsPutAndSubscriptionForCommitOnIndexFailsThenTheChannelIsUnblocked(t *testing.T) {

	testContext := setup()

	testContext.journalSpy.FailNextNotifyOfCommitOnIndexOnce()

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	result := <-doneCh

	if result {
		t.Errorf("Channel should have been unblocked with false value because subscription on commit index failed")
	}
}

func TestWhenEmptyKeyIsPutThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	_, err := testContext.client.Put("", testContext.data)

	if ErrEmptyKey != err {

		t.Errorf("Expected error '%v', but got '%v'", ErrEmptyKey, err)
	}
}

func TestWhenEmptyKeyIsPutThenTheChannelIsUnblocked(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put("", testContext.data)

	result := <-doneCh

	if result {

		t.Errorf("Key is invalid, channel should have read false")
	}
}

func TestWhenGettingDataForKeyThatDoesNotExistThenAnErrorIsReturned(t *testing.T) {

	testContext := setup()

	_, getErr := testContext.client.Get("blah")

	if ErrKeyDoesNotExist != getErr {
		t.Errorf("Should have recieved error '%v' but got '%v'",
			ErrKeyDoesNotExist,
			getErr)
	}
}

//func TestWhenGettingDataForKeyThatExistsThenTheDataIsReturned(t *testing.T) {
//
//	testContext := setup()
//
//	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)
//
//	<-doneCh
//
//	retrievedData, _ := testContext.client.Get(testContext.key)
//
//	if !reflect.DeepEqual(testContext.data, retrievedData) {
//
//		t.Errorf("Data should have been retrieved for key '%s', expected '%s' but got '%s'",
//			testContext.key,
//			testContext.data,
//			retrievedData)
//	}
//}

// ============================= Test Support ==================

type TestContext struct {
	client     *Client
	journalSpy *journal.Spy
	key        string
	data       []byte
}

func hashIt(key string) []byte {

	hashFunc := sha1.New()

	hash := hashFunc.Sum([]byte(key))

	return hash
}

func setup() *TestContext {

	journalSpy := journal.New()

	client := New(journalSpy)

	testContext := &TestContext{
		client:     client,
		journalSpy: journalSpy,
		key:        "iamakey",
		data:       []byte("i am some data"),
	}

	return testContext
}
