package bootstrap

import (
	"github.com/jrobison153/raft/journal"
	"github.com/jrobison153/raft/state"
	"reflect"
	"strings"
	"testing"
)

func TestWhenCreatingPolicyClientThenCorrectDefaultLoggerIsUsed(t *testing.T) {

	server := Init()

	typeOfExpectedLogger := reflect.TypeOf(&journal.ArrayJournal{}).String()
	typeOfActualLogger := server.PolicyClient().TypeOfLogger()

	if strings.Compare(typeOfExpectedLogger, typeOfActualLogger) != 0 {
		t.Errorf("Default logger was not set, expected logger type %v but got %v",
			typeOfExpectedLogger,
			typeOfActualLogger)
	}
}

func TestWhenCreatingPolicyClientThenCorrectDefaultStateMachineIsUsed(t *testing.T) {

	server := Init()

	typeOfExpectedStateMachine := reflect.TypeOf(&state.MapStateMachine{}).String()
	typeOfActualStateMachine := server.PolicyClient().TypeOfStateMachine()

	if strings.Compare(typeOfExpectedStateMachine, typeOfActualStateMachine) != 0 {
		t.Errorf("Default state machine was not set, expected state machine type %v but got %v",
			typeOfExpectedStateMachine,
			typeOfActualStateMachine)
	}
}
