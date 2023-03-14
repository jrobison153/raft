package state

import (
	"errors"
	"github.com/jrobison153/raft/journal"
	"reflect"
)

type Spy struct {
	renderedState [][]byte
	journal       journal.Journaler
	startCalled   bool
}

func NewStateMachineSpy(journal journal.Journaler) *Spy {
	return &Spy{
		renderedState: make([][]byte, 0, 16),
		journal:       journal,
	}
}

func (spy *Spy) ResolveRequestToData(request []byte) ([]byte, error) {

	var data []byte
	data = spy.retrieveItemFromStore(request, data)

	var err error
	if data == nil {
		err = errors.New("failing for test purposes, requested data does not exist")
	}

	return data, err
}

func (spy *Spy) retrieveItemFromStore(request []byte, data []byte) []byte {
	for _, storedData := range spy.renderedState {

		if reflect.DeepEqual(request, storedData) {
			data = storedData
			break
		}
	}
	return data
}

func (spy *Spy) Start() error {
	spy.startCalled = true
	return nil
}

func (spy *Spy) TypeOfLogger() string {

	return reflect.TypeOf(spy.journal).String()
}

func (spy *Spy) StartCalled() bool {
	return spy.startCalled
}

func (spy *Spy) AddStateMachineData(item []byte) {

	spy.renderedState = append(spy.renderedState, item)
}
