package state

import "errors"

type Spy struct {
	renderedState map[string][]byte
}

func NewStateMachineSpy() *Spy {
	return &Spy{
		renderedState: make(map[string][]byte),
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

func (spy *Spy) SetDataForKey(key string, data []byte) {

	spy.renderedState[key] = data
}
