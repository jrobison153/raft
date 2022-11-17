package state

import (
	"errors"
	"github.com/jrobison153/raft/journal"
	"reflect"
)

type Spy struct {
	renderedState map[string][]byte
	journal       journal.Journaler
	startCalled   bool
}

func NewStateMachineSpy(journal journal.Journaler) *Spy {
	return &Spy{
		renderedState: make(map[string][]byte),
		journal:       journal,
	}
}

func (spy *Spy) GetValueForKey(key string) ([]byte, error) {

	data := spy.renderedState[key]

	var err error
	if data == nil {
		err = errors.New("failing for test purposes, key does not exist in state")
	}

	return spy.renderedState[key], err
}

func (spy *Spy) Start() error {
	spy.startCalled = true
	return nil
}

func (spy *Spy) TypeOfLogger() string {

	return reflect.TypeOf(spy.journal).String()
}

func (spy *Spy) SetDataForKey(key string, data []byte) {

	spy.renderedState[key] = data
}

func (spy *Spy) StartCalled() bool {
	return spy.startCalled
}
