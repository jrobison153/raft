package replication

import (
	"github.com/jrobison153/raft/journal"
	"reflect"
)

type Spy struct {
	startCalled bool
	journal     journal.Journaler
	startTimer  Timer
}

func NewReplicatorSpy(journal journal.Journaler) *Spy {

	return &Spy{
		startCalled: false,
		journal:     journal,
	}
}

func (spy *Spy) TypeOfLogger() string {

	return reflect.TypeOf(spy.journal).String()
}

func (spy *Spy) StartCalled() bool {

	return spy.startCalled
}

func (spy *Spy) Start(timer Timer) {

	spy.startTimer = timer
	spy.startCalled = true
}

func (spy *Spy) TypeOfStartTimer() string {

	return reflect.TypeOf(spy.startTimer).String()
}
