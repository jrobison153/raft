package grpc

import (
	"context"
	"errors"
	"github.com/jrobison153/raft/api"
	"github.com/jrobison153/raft/policy/client"
	"reflect"
	"strings"
	"testing"
)

const expectedKey = "12345678"

var expectedData = []byte("Data here! Just passing through!")

const (
	NoFailures = iota
	GeneralFailure
	ReplicationFailure
)

func TestWhenItemIsPutThenTheKeyIsHandedToTheSupportingPolicyService(t *testing.T) {

	testContext := setupForPut(NoFailures)

	key := testContext.ClientPolicySpy.LastPutKey

	if strings.Compare(expectedKey, key()) != 0 {
		t.Errorf("Key should have been handed to underlying policy service, expected '%s' but saw '%s'",
			expectedKey,
			key())

	}
}

func TestWhenItemIsPutThenTheDataIsHandedToTheSupportingPolicyService(t *testing.T) {

	testContext := setupForPut(NoFailures)

	key := testContext.ClientPolicySpy.LastDataKey()

	if !reflect.DeepEqual(expectedData, key) {

		t.Errorf("Data should have been handed to underlying policy service, expected '%s' but saw '%s'",
			expectedData,
			key)
	}
}

func TestWhenPutCompletesSuccessfullyThenStatusIsOk(t *testing.T) {

	testContext := setupForPut(NoFailures)

	actualStatus := testContext.PutResponse.Status

	if api.ClientStatusCodes_PUT_OK != actualStatus {

		t.Errorf("PutResponse status should have been %s but instead we got %s",
			api.ClientStatusCodes_PUT_OK,
			actualStatus)
	}
}

func TestWhenPutCompletesSuccessfullyThenRetryStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForPut(NoFailures)

	isRetryable := testContext.PutResponse.IsRetryable
	if api.RetryCodes_NO != isRetryable {

		t.Errorf("PutItem should have returned with retryable code '%v', instead we got '%v'",
			api.RetryCodes_NO,
			isRetryable)
	}
}

func TestWhenPutFailsWithAnErrorThenTheCorrectErrorIsReturned(t *testing.T) {

	testContext := setupForPut(GeneralFailure)

	putErr := testContext.PutErr

	if !errors.Is(ErrPutRetryable, putErr) {
		t.Errorf("PutItem should have returned an error '%v', instead we got '%v'",
			ErrPutRetryable,
			putErr)
	}
}

func TestWhenPutFailsWithAnErrorThenTheStatusIsPutError(t *testing.T) {

	testContext := setupForPut(GeneralFailure)

	status := testContext.PutResponse.Status

	if api.ClientStatusCodes_PUT_ERROR != status {
		t.Errorf("PutItem should have returned with Status '%v', instead we got '%v'",
			api.ClientStatusCodes_PUT_ERROR,
			status)
	}
}

func TestWhenPutFailsWithAnErrorAndItIsRetryableThenTheRetryableStatusIsSetToYes(t *testing.T) {

	testContext := setupForPut(GeneralFailure)

	isRetryable := testContext.PutResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {
		t.Errorf("PutItem should have returned with retryable code '%v', instead we got '%v'",
			api.RetryCodes_YES,
			isRetryable)
	}
}

func TestWhenPutCompletesSuccessfullyThenReplicationCompleteStatusIsSetToSuccessValue(t *testing.T) {

	testContext := setupForPut(NoFailures)

	replStatus := testContext.PutResponse.ReplicationStatus

	if api.ReplicationCodes_QUORUM_REACHED != replStatus {
		t.Errorf("PutItem should have returned replication status of '%v' but instead we got '%v'",
			api.ReplicationCodes_QUORUM_REACHED,
			replStatus)
	}
}

func TestWhenPutReplicationFailsThenReplicationCompleteStatusIsSetToFailureValue(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	replStatus := testContext.PutResponse.ReplicationStatus

	if api.ReplicationCodes_FAILURE_TO_REACH_QUORUM != replStatus {
		t.Errorf("PutItem should have returned replication status of '%v' but instead we got '%v'",
			api.ReplicationCodes_FAILURE_TO_REACH_QUORUM,
			replStatus)
	}
}

func TestWhenPutReplicationFailsThenTheStatusIsPutError(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	status := testContext.PutResponse.Status

	if api.ClientStatusCodes_PUT_ERROR != status {
		t.Errorf("PutItem should have returned put status of '%v' but instead we got '%v'",
			api.ClientStatusCodes_PUT_ERROR,
			status)
	}
}

func TestWhenPutReplicationFailsThenTheRetryStatusIsSetToYes(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	isRetryable := testContext.PutResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {
		t.Errorf("PutResponse restyrable status should have been '%v' but was '%v'",
			api.RetryCodes_YES,
			isRetryable)
	}
}

func TestWhenPutReplicationFailsThenTheCorrectErrorIsReturned(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	putErr := testContext.PutErr

	if ErrPutRetryable != putErr {
		t.Errorf("Expected error '%v' but got '%v'", ErrPutRetryable, putErr)
	}
}

func TestWhenGettingItemThatDoesNotExistThenAnErrorIsReturned(t *testing.T) {

	testContext := setupForGet("non-existent key")

	if testContext.GetErr == nil {
		t.Errorf("An error should have been returned, but it was not")
	}
}

func TestWhenGettingItemThatDoesNotExistThenTheCorrectErrorIsReturned(t *testing.T) {

	testContext := setupForGet("non-existent key")

	if ErrGetNonExistentKey != testContext.GetErr {
		t.Errorf("Should have seen error '%v', but instead got '%v'",
			ErrGetNonExistentKey,
			testContext.GetErr)
	}
}

func TestWhenGettingItemThatDoesNotExistThenResponseStatusSetCorrectly(t *testing.T) {

	testContext := setupForGet("non-existent key")

	status := testContext.GetResponse.Status

	if api.ClientStatusCodes_GET_ERROR != status {

		t.Errorf("PutItem should have returned put status of '%v' but instead we got '%v'",
			api.ClientStatusCodes_GET_ERROR,
			status)
	}
}

func TestWhenGettingItemThatDoesNotExistThenRetryStatusSetCorrectly(t *testing.T) {

	testContext := setupForGet("non-existent key")

	isRetryable := testContext.GetResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {

		t.Errorf("PutItem should have returned retry status of '%v' but instead we got '%v'",
			api.RetryCodes_YES,
			isRetryable)
	}
}

func TestWhenGettingItemThatDoesNotExistThenErrorMessageSetCorrectly(t *testing.T) {

	testContext := setupForGet("non-existent key")

	errorMessage := testContext.GetResponse.ErrorMessage

	if strings.Compare(ErrGetNonExistentKey.Error(), errorMessage) != 0 {

		t.Errorf("PutItem should have returned error message '%s' but instead we got '%s'",
			ErrGetNonExistentKey.Error(),
			errorMessage)
	}
}

func TestWhenGettingItemWithEmptyKeyThenTheCorrectErrorIsReturned(t *testing.T) {

	testContext := setupForGet("")

	if ErrGetEmptyKey != testContext.GetErr {
		t.Errorf("Should have seen error '%v', but instead got '%v'",
			ErrGetEmptyKey,
			testContext.GetErr)
	}
}

func TestWhenGettingItemWithEmptyKeyThenTheResponseStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForGet("")

	status := testContext.GetResponse.GetStatus()

	if api.ClientStatusCodes_GET_ERROR != status {
		t.Errorf("Should have got status '%v', but instead got '%v'",
			api.ClientStatusCodes_GET_ERROR,
			status)
	}
}

func TestWhenGettingItemWithEmptyKeyThenTheCorrectErrorMessageIsReturned(t *testing.T) {

	testContext := setupForGet("")

	errorMessage := testContext.GetResponse.ErrorMessage

	if strings.Compare(ErrGetEmptyKey.Error(), errorMessage) != 0 {

		t.Errorf("Should seen error message '%s', but instead got '%s'",
			ErrGetEmptyKey,
			errorMessage)
	}
}

func TestWhenGettingItemWithEmptyKeyThenTheCorrectRetryStatusIsReturned(t *testing.T) {

	testContext := setupForGet("")

	isRetryable := testContext.GetResponse.IsRetryable

	if api.RetryCodes_NO != isRetryable {
		t.Errorf("Retry status should have been '%v' but instead was '%v'",
			api.RetryCodes_NO,
			isRetryable)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheKeyIsReturnedInTheResponse(t *testing.T) {

	testContext := setupForGet(expectedKey)

	if strings.Compare(expectedKey, testContext.GetResponse.Item.Key) != 0 {

		t.Errorf("Key shoud have been returned in the response but it was not, saw response '%+v'",
			testContext.GetResponse)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheDataIsReturnedInTheResponse(t *testing.T) {

	testContext := setupForGet(expectedKey)

	if !reflect.DeepEqual(expectedData, testContext.GetResponse.Item.Data) {

		t.Errorf("Data shoud have been returned in the response but it was not, saw response '%+v'",
			testContext.GetResponse)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheResponseStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForGet(expectedKey)

	status := testContext.GetResponse.Status

	if api.ClientStatusCodes_GET_OK != status {
		t.Errorf("Response should have had status %v but we got %v",
			api.ClientStatusCodes_GET_OK,
			status)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheRetryStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForGet(expectedKey)

	isRetryable := testContext.GetResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {
		t.Errorf("Response should have had retry status %v but we got %v",
			api.RetryCodes_YES,
			isRetryable)
	}
}

func setup(spy *client.Spy) *RaftServer {

	server := New(spy)

	return server
}

type TestContext struct {
	PutResponse     *api.PutResponse
	GetResponse     *api.GetResponse
	ClientPolicySpy *client.Spy
	PutErr          error
	GetErr          error
	GetKey          *api.Key
}

func setupForPut(scenario int) *TestContext {

	clientPolicySpy := client.NewSpy()

	if scenario == GeneralFailure {
		clientPolicySpy.FailNextPut()
	}

	if scenario == ReplicationFailure {
		clientPolicySpy.FailReplication()
	}

	server := setup(clientPolicySpy)

	item := &api.Item{
		Key:  expectedKey,
		Data: expectedData,
	}

	response, putErr := server.PutItem(context.Background(), item)

	testContext := &TestContext{
		PutResponse:     response,
		ClientPolicySpy: clientPolicySpy,
		PutErr:          putErr,
	}

	return testContext
}

func setupForGet(aKey string) *TestContext {

	spy := client.NewSpy()

	server := setup(spy)

	key := &api.Key{
		Key: aKey,
	}

	item := &api.Item{
		Key:  expectedKey,
		Data: expectedData,
	}

	server.PutItem(context.Background(), item)

	response, err := server.GetItem(context.Background(), key)

	testContext := &TestContext{
		ClientPolicySpy: spy,
		GetKey:          key,
		GetErr:          err,
		GetResponse:     response,
	}

	return testContext
}
