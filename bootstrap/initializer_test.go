package bootstrap

import (
	"github.com/jrobison153/raft/journal"
	"github.com/jrobison153/raft/replication"
	"github.com/jrobison153/raft/server"
	"github.com/jrobison153/raft/state"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestWhenCreatingPolicyClientThenCorrectDefaultLoggerIsUsed(t *testing.T) {

	bootstrapper := New()
	_ = bootstrapper.Init()

	typeOfExpectedLogger := reflect.TypeOf(&journal.ArrayJournal{}).String()
	typeOfActualLogger := bootstrapper.clientPolicy.TypeOfLogger()

	if strings.Compare(typeOfExpectedLogger, typeOfActualLogger) != 0 {
		t.Errorf("Default logger was not set, expected logger type %v but got %v",
			typeOfExpectedLogger,
			typeOfActualLogger)
	}
}

func TestWhenCreatingPolicyClientThenCorrectDefaultStateMachineIsUsed(t *testing.T) {

	bootstrapper := New()
	_ = bootstrapper.Init()

	typeOfExpectedStateMachine := reflect.TypeOf(&state.MapStateMachine{}).String()
	typeOfActualStateMachine := bootstrapper.clientPolicy.TypeOfStateMachine()

	if strings.Compare(typeOfExpectedStateMachine, typeOfActualStateMachine) != 0 {
		t.Errorf("Default state machine was not set, expected state machine type %v but got %v",
			typeOfExpectedStateMachine,
			typeOfActualStateMachine)
	}
}

func TestWhenCreatingPolicyClientThenCorrectDefaultReplicatorIsUsed(t *testing.T) {

	bootstrapper := New()
	_ = bootstrapper.Init()

	typeOfExpectedReplicator := reflect.TypeOf(&replication.NoOpReplicator{}).String()
	typeOfActualReplicator := reflect.TypeOf(bootstrapper.replicator).String()

	if strings.Compare(typeOfExpectedReplicator, typeOfActualReplicator) != 0 {
		t.Errorf("Default replicator was not set, expected replicator type %v but got %v",
			typeOfExpectedReplicator,
			typeOfActualReplicator)
	}
}

func TestWhenCreatingNoOpReplicatorThenItIsConstructedWithTheDefaultConfig(t *testing.T) {

	bootstrapper := New()
	_ = bootstrapper.Init()

	noOpReplicator := bootstrapper.replicator.(*replication.NoOpReplicator)
	spiedDefaultConfig := noOpReplicator.ReplicatorConfig()

	expectedDefaultConfig := *(replication.NewDefaultConfig())

	if !reflect.DeepEqual(expectedDefaultConfig, spiedDefaultConfig) {
		t.Errorf("Replicator should have been initialized with default config %+v but instead got %+v",
			expectedDefaultConfig,
			spiedDefaultConfig)
	}
}

func TestWhenJournalTypeEnvVarIsSetToAnUnknownValueThenAnErrorIsReturned(t *testing.T) {

	_ = os.Setenv(journalTypeEnvVar, "bogus")
	defer envCleanUp(journalTypeEnvVar)

	bootstrapper := New()
	err := bootstrapper.Init()

	if err != ErrInvalidJournalType {
		t.Errorf("Should have received error '%v' but instead got '%v'", ErrInvalidJournalType, err)
	}
}

func TestWhenStateMachineTypeEnvVarIsSetToAnUnknownValueThenAnErrorIsReturned(t *testing.T) {

	_ = os.Setenv(stateMachineTypeEnvVar, "bogus")
	defer envCleanUp(stateMachineTypeEnvVar)

	bootstrapper := New()
	err := bootstrapper.Init()

	if err != ErrInvalidStateMachineType {
		t.Errorf("Should have received error '%v' but instead got '%v'", ErrInvalidStateMachineType, err)
	}
}

func TestWhenReplicatorTypeEnvVarIsSetToAnUnknownValueThenAnErrorIsReturned(t *testing.T) {

	_ = os.Setenv(replicatorTypeEnvVar, "bogus")
	defer envCleanUp(replicatorTypeEnvVar)

	bootstrapper := New()
	err := bootstrapper.Init()

	if err != ErrInvalidReplicatorType {
		t.Errorf("Should have received error '%v' but instead got '%v'", ErrInvalidReplicatorType, err)
	}
}

func TestWhenUnableToCreateJournalThenTheStateMachineIsNotCreated(t *testing.T) {

	_ = os.Setenv(journalTypeEnvVar, "bogus")
	defer envCleanUp(journalTypeEnvVar)

	bootstrapper := New()
	bootstrapper.Init()

	if bootstrapper.stateMachine != nil {
		t.Errorf("State machine should not have been initialized due to previous errors")
	}
}

func TestWhenUnableToCreateJournalThenTheReplicatorIsNotCreated(t *testing.T) {

	_ = os.Setenv(journalTypeEnvVar, "bogus")
	defer envCleanUp(journalTypeEnvVar)

	bootstrapper := New()
	bootstrapper.Init()

	if bootstrapper.replicator != nil {
		t.Errorf("Replicator should not have been initialized due to previous errors")
	}
}

func TestWhenBootstrappingThenTheStateMachineIsCreatedWithTheLog(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	typeOfExpectedLogger := reflect.TypeOf(&journal.Spy{}).String()
	typeOfActualLogger := bootstrapper.stateMachine.TypeOfLogger()

	if strings.Compare(typeOfExpectedLogger, typeOfActualLogger) != 0 {
		t.Errorf("State machine was not passed the logger on creation, expected logger type %v but got %v",
			typeOfExpectedLogger,
			typeOfActualLogger)
	}
}

func TestWhenBootstrappingThenTheReplicatorIsCreatedWithTheLog(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	typeOfExpectedLogger := reflect.TypeOf(&journal.Spy{}).String()
	typeOfActualLogger := bootstrapper.replicator.TypeOfLogger()

	if strings.Compare(typeOfExpectedLogger, typeOfActualLogger) != 0 {
		t.Errorf("Replicator was not passed the logger on creation, expected logger type '%v' but got '%v'",
			typeOfExpectedLogger,
			typeOfActualLogger)
	}
}

func TestWhenJournalTypeEnvVarSetToSpyThenTheSpyJournalImplementationIsUsed(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	expectedJournalType := reflect.TypeOf(&journal.Spy{}).String()
	actualJournalType := reflect.TypeOf(bootstrapper.journal).String()

	if strings.Compare(expectedJournalType, actualJournalType) != 0 {
		t.Errorf("Journal should have been set to type '%v' but instead was '%v'",
			expectedJournalType,
			actualJournalType)
	}
}

func TestWhenStateMachineTypeEnvVarSetToSpyThenTheSpyStateMachineImplementationIsUsed(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	expectedStateMachineType := reflect.TypeOf(&state.Spy{}).String()
	actualStateMachineType := reflect.TypeOf(bootstrapper.stateMachine).String()

	if strings.Compare(expectedStateMachineType, actualStateMachineType) != 0 {
		t.Errorf("State machine should have been set to type '%v' but instead was '%v'",
			expectedStateMachineType,
			actualStateMachineType)
	}
}

func TestWhenUnableToCreateJournalThenTheServerIsNotCreated(t *testing.T) {

	_ = os.Setenv(journalTypeEnvVar, "bogus")
	defer envCleanUp(journalTypeEnvVar)

	bootstrapper := New()
	_ = bootstrapper.Init()

	if bootstrapper.server != nil {
		t.Errorf("Server should not have been initialized due to previous errors")
	}
}

func TestWhenUnableToCreateTheRendererThenTheServerIsNotCreated(t *testing.T) {

	_ = os.Setenv(replicatorTypeEnvVar, "bogus")
	defer envCleanUp(replicatorTypeEnvVar)

	bootstrapper := New()
	_ = bootstrapper.Init()

	if bootstrapper.server != nil {
		t.Errorf("Server should not have been initialized due to previous errors")
	}
}

func TestWhenUnableToCreateStateMachineThenServerIsNotCreated(t *testing.T) {

	_ = os.Setenv(journalTypeEnvVar, spyJournalType)
	_ = os.Setenv(stateMachineTypeEnvVar, "bogus")
	defer envCleanUp(journalTypeEnvVar)
	defer envCleanUp(stateMachineTypeEnvVar)

	bootstrapper := New()
	bootstrapper.Init()

	if bootstrapper.server != nil {
		t.Errorf("Server should not have been initialized due to previous errors")
	}
}

func TestWhenStartingTheSystemThenTheStateMachineIsStarted(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	bootstrapper.Start(0)

	stateMachineSpy := bootstrapper.stateMachine.(*state.Spy)

	if !stateMachineSpy.StartCalled() {
		t.Errorf("State machine should have been started")
	}
}

func TestWhenStartingTheSystemThenTheReplicatorIsStarted(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	bootstrapper.Start(0)

	replicatorSpy := bootstrapper.replicator.(*replication.Spy)

	if !replicatorSpy.StartCalled() {
		t.Errorf("State machine should have been started")
	}
}

func TestWhenStartingTheSystemThenTheReplicatorIsStartedTheDefaultTimer(t *testing.T) {

	bootstrapper := setup()
	defer tearDown()

	bootstrapper.Start(0)

	replicatorSpy := bootstrapper.replicator.(*replication.Spy)

	expectedTimerType := reflect.TypeOf(&replication.SleepTimer{}).String()
	actualTimerType := replicatorSpy.TypeOfStartTimer()

	if strings.Compare(expectedTimerType, actualTimerType) != 0 {
		t.Errorf("State machine should have been started with timer %v but got %v",
			expectedTimerType,
			actualTimerType)
	}
}

func TestWhenStartingTheSystemThenTheServerIsStartedOnProvidedPort(t *testing.T) {

	configEnvForSpies()
	defer unsetSpiedEnv()

	bootstrapper := New()
	_ = bootstrapper.Init()

	serverSpy := server.NewLifeCyclerSpy()
	bootstrapper.server = serverSpy

	const port = 9999
	bootstrapper.Start(port)

	if !serverSpy.StartCalledOnPort(port) {
		t.Errorf("Client API server should have been started on port %d", port)
	}
}

func setup() *Bootstrap {

	configEnvForSpies()

	bootstrapper := New()
	_ = bootstrapper.Init()

	serverSpy := server.NewLifeCyclerSpy()
	bootstrapper.server = serverSpy

	return bootstrapper
}

func tearDown() {
	unsetSpiedEnv()
}

func envCleanUp(envVarName string) {

	_ = os.Unsetenv(envVarName)
}

func unsetSpiedEnv() {
	envCleanUp(journalTypeEnvVar)
	envCleanUp(stateMachineTypeEnvVar)
	envCleanUp(replicatorTypeEnvVar)
}

func configEnvForSpies() {

	_ = os.Setenv(journalTypeEnvVar, spyJournalType)
	_ = os.Setenv(stateMachineTypeEnvVar, spyStateMachineType)
	_ = os.Setenv(replicatorTypeEnvVar, spyReplicatorType)
}
