package client

import (
	"crypto/sha1"
	"github.com/jrobison153/raft/replicator"
	"github.com/jrobison153/raft/test/doubles/log"
	replSpy "github.com/jrobison153/raft/test/doubles/replicator"
	"reflect"
	"strings"
	"testing"
)

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

	if testContext.logSpy.AppendCalled() != true {
		t.Error("Put item was not appended to the log")
	}
}

func TestWhenItemIsPutThenTheKeyIsHashed(t *testing.T) {

	testContext := setup()

	testContext.client.Put(testContext.key, testContext.data)

	spiedKey := testContext.logSpy.AppendKey()
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

	spiedData := testContext.logSpy.LatestDataForKey(hashedKey)
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
	testContext.logSpy.FailNextAppend("whatever")

	_, err := testContext.client.Put(testContext.key, testContext.data)

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestWhenLogAppendFailsThenTheCorrectErrorMessageIsReturned(t *testing.T) {

	testContext := setup()
	testContext.logSpy.FailNextAppend("failing for test purposes")

	_, err := testContext.client.Put(testContext.key, testContext.data)

	expectedErrMsg := "client: append to log failed, root cause 'failing for test purposes'"

	if strings.Compare(err.Error(), expectedErrMsg) != 0 {

		t.Errorf("Expected error '%s' but got non-matching error '%s'",
			expectedErrMsg,
			err.Error())
	}
}

func TestWhenItemIsPutThenItIsReplicated(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	<-doneCh

	expectedReplicatedItem := replicator.Item{
		Key:  hashIt(testContext.key),
		Data: testContext.data,
	}

	replicatedItem := testContext.replicatorSpy.ReplicatedItem()

	if !reflect.DeepEqual(expectedReplicatedItem, *replicatedItem) {
		t.Errorf("Expected key %+v to be replicated but actually replicated %+v",
			expectedReplicatedItem,
			replicatedItem)
	}
}

func TestWhenItemIPutThenChannelReturnedUnblocksOnSuccess(t *testing.T) {

	testContext := setup()

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	isPutComplete := <-doneCh

	if !isPutComplete {
		t.Error("Expected replication to complete successfully")
	}
}

func TestWhenLogAppendFailThenReplicationDoesNotHappen(t *testing.T) {

	testContext := setup()

	testContext.logSpy.FailNextAppend("Failing for test purposes")

	doneCh, _ := testContext.client.Put(testContext.key, testContext.data)

	<-doneCh

	if testContext.replicatorSpy.ReplicateCalled() {
		t.Error("An attempt was made to replicate the item when it should not have been")
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

		t.Errorf("Key is invalid, replication channel should have read false")
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
	client        *Client
	logSpy        *log.Spy
	replicatorSpy *replSpy.Spy
	key           string
	data          []byte
}

func hashIt(key string) []byte {

	hashFunc := sha1.New()

	hash := hashFunc.Sum([]byte(key))

	return hash
}

func setup() *TestContext {

	logSpy := log.New()
	replicatorSpy := replSpy.New()

	client := New(logSpy, replicatorSpy)

	testContext := &TestContext{
		client:        client,
		logSpy:        logSpy,
		replicatorSpy: replicatorSpy,
		key:           "iamakey",
		data:          []byte("i am some data"),
	}

	return testContext
}
