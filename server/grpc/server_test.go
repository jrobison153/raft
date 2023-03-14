package grpc

import (
	"context"
	"errors"
	"github.com/jrobison153/raft/api"
	"github.com/jrobison153/raft/policy/client"
	"strings"
	"testing"
)

var expectedData = "Item here! Just passing through!"

const (
	NoFailures = iota
	GeneralFailure
	ReplicationFailure
)

func TestWhenItemIsPutThenTheItemIsHandedToTheSupportingPolicyService(t *testing.T) {

	testContext := setupForPut(NoFailures)

	item := testContext.ClientPolicySpy.LastPutItem()

	itemAsStr := string(item)

	if strings.Compare(expectedData, itemAsStr) != 0 {
		t.Errorf("Put item should have been handed to underlying policy service, expected '%s' but saw '%s'",
			expectedData,
			itemAsStr)

	}
}

func TestWhenPutCompletesSuccessfullyThenStatusIsOk(t *testing.T) {

	testContext := setupForPut(NoFailures)

	actualStatus := testContext.PutItemResponse.Status

	if api.ClientStatusCodes_PUT_OK != actualStatus {

		t.Errorf("PutItemResponse status should have been %s but instead we got %s",
			api.ClientStatusCodes_PUT_OK,
			actualStatus)
	}
}

func TestWhenPutCompletesSuccessfullyThenRetryStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForPut(NoFailures)

	isRetryable := testContext.PutItemResponse.IsRetryable
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

	status := testContext.PutItemResponse.Status

	if api.ClientStatusCodes_PUT_ERROR != status {
		t.Errorf("PutItem should have returned with Status '%v', instead we got '%v'",
			api.ClientStatusCodes_PUT_ERROR,
			status)
	}
}

func TestWhenPutFailsWithAnErrorAndItIsRetryableThenTheRetryableStatusIsSetToYes(t *testing.T) {

	testContext := setupForPut(GeneralFailure)

	isRetryable := testContext.PutItemResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {
		t.Errorf("PutItem should have returned with retryable code '%v', instead we got '%v'",
			api.RetryCodes_YES,
			isRetryable)
	}
}

func TestWhenPutCompletesSuccessfullyThenReplicationCompleteStatusIsSetToSuccessValue(t *testing.T) {

	testContext := setupForPut(NoFailures)

	replStatus := testContext.PutItemResponse.ReplicationStatus

	if api.ReplicationCodes_QUORUM_REACHED != replStatus {
		t.Errorf("PutItem should have returned replication status of '%v' but instead we got '%v'",
			api.ReplicationCodes_QUORUM_REACHED,
			replStatus)
	}
}

func TestWhenPutReplicationFailsThenReplicationCompleteStatusIsSetToFailureValue(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	replStatus := testContext.PutItemResponse.ReplicationStatus

	if api.ReplicationCodes_FAILURE_TO_REACH_QUORUM != replStatus {
		t.Errorf("PutItem should have returned replication status of '%v' but instead we got '%v'",
			api.ReplicationCodes_FAILURE_TO_REACH_QUORUM,
			replStatus)
	}
}

func TestWhenPutReplicationFailsThenTheStatusIsPutError(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	status := testContext.PutItemResponse.Status

	if api.ClientStatusCodes_PUT_ERROR != status {
		t.Errorf("PutItem should have returned put status of '%v' but instead we got '%v'",
			api.ClientStatusCodes_PUT_ERROR,
			status)
	}
}

func TestWhenPutReplicationFailsThenTheRetryStatusIsSetToYes(t *testing.T) {

	testContext := setupForPut(ReplicationFailure)

	isRetryable := testContext.PutItemResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {
		t.Errorf("PutItemResponse restyrable status should have been '%v' but was '%v'",
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

func TestWhenGettingItemAndThereIsAnErrorThenTheCorrectErrorIsReturned(t *testing.T) {

	testContext := setupForGet(GeneralFailure)

	if ErrGetGeneralFailure != testContext.GetErr {
		t.Errorf("Should have seen error '%v', but instead got '%v'",
			ErrGetGeneralFailure,
			testContext.GetErr)
	}
}

func TestWhenGettingItemAndThereIsAnErrorThenResponseStatusSetCorrectly(t *testing.T) {

	testContext := setupForGet(GeneralFailure)

	status := testContext.GetItemResponse.Status

	if api.ClientStatusCodes_GET_ERROR != status {

		t.Errorf("PutItem should have returned put status of '%v' but instead we got '%v'",
			api.ClientStatusCodes_GET_ERROR,
			status)
	}
}

func TestWhenGettingItemAndThereIsAnErrorThenRetryStatusSetCorrectly(t *testing.T) {

	testContext := setupForGet(GeneralFailure)

	isRetryable := testContext.GetItemResponse.IsRetryable

	if api.RetryCodes_YES != isRetryable {

		t.Errorf("PutItem should have returned retry status of '%v' but instead we got '%v'",
			api.RetryCodes_YES,
			isRetryable)
	}
}

func TestWhenGettingItemAndThereIsAnErrorThenErrorMessageSetCorrectly(t *testing.T) {

	testContext := setupForGet(GeneralFailure)

	errorMessage := testContext.GetItemResponse.ErrorMessage

	if strings.Compare(ErrGetGeneralFailure.Error(), errorMessage) != 0 {

		t.Errorf("PutItem should have returned error message '%s' but instead we got '%s'",
			ErrGetGeneralFailure.Error(),
			errorMessage)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheItemIsReturnedInTheResponse(t *testing.T) {

	testContext := setupForGet(NoFailures)

	responseData := string(testContext.GetItemResponse.Item.Data)

	if strings.Compare(expectedData, responseData) != 0 {

		t.Errorf("Item should have been returned in the response but it was not, saw response '%+v'",
			testContext.GetItemResponse)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheResponseStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForGet(NoFailures)

	status := testContext.GetItemResponse.Status

	if api.ClientStatusCodes_GET_OK != status {
		t.Errorf("Response should have had status %v but we got %v",
			api.ClientStatusCodes_GET_OK,
			status)
	}
}

func TestWhenGettingItemThatHasBeenPreviouslyPutThenTheRetryStatusIsSetCorrectly(t *testing.T) {

	testContext := setupForGet(NoFailures)

	isRetryable := testContext.GetItemResponse.IsRetryable

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
	PutItemResponse *api.PutItemResponse
	GetItemResponse *api.GetItemResponse
	ClientPolicySpy *client.Spy
	PutErr          error
	GetErr          error
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
		Data: []byte(expectedData),
	}

	response, putErr := server.PutItem(context.Background(), item)

	testContext := &TestContext{
		PutItemResponse: response,
		ClientPolicySpy: clientPolicySpy,
		PutErr:          putErr,
	}

	return testContext
}

func setupForGet(scenario int) *TestContext {

	spy := client.NewSpy()

	server := setup(spy)

	if scenario == GeneralFailure {
		spy.FailNextGet()
	}

	item := &api.Item{
		Data: []byte(expectedData),
	}

	server.PutItem(context.Background(), item)

	response, err := server.GetItem(context.Background(), item)

	testContext := &TestContext{
		ClientPolicySpy: spy,
		GetErr:          err,
		GetItemResponse: response,
	}

	return testContext
}
