package client

import (
	"errors"
)

type Spy struct {
	lastPutItem          []byte
	isFailingPut         bool
	isFailingReplication bool
	isFailingGet         bool
}

func NewSpy() *Spy {

	return &Spy{
		isFailingPut:         false,
		isFailingReplication: false,
	}
}

func (spy *Spy) TypeOfLogger() string {
	//TODO implement me
	panic("implement me")
}

func (spy *Spy) TypeOfStateMachine() string {
	//TODO implement me
	panic("implement me")
}

func (spy *Spy) Put(item []byte) (chan bool, error) {

	var err error
	var replResult = true

	if spy.isFailingPut {

		err = errors.New("failing Put for test reasons")
	} else if spy.isFailingReplication {
		replResult = false
	} else {

		spy.lastPutItem = item
	}

	doneCh := make(chan bool)

	go func(ch chan bool, val bool) { ch <- replResult }(doneCh, replResult)

	return doneCh, err
}

func (spy *Spy) Get(item []byte) ([]byte, error) {

	var err error
	var data []byte

	if spy.isFailingGet {
		err = errors.New("failing Get for test reasons")
	} else {
		data = spy.lastPutItem
	}

	return data, err
}

func (spy *Spy) LastPutItem() []byte {

	return spy.lastPutItem
}

func (spy *Spy) FailNextPut() {
	spy.isFailingPut = true
}

func (spy *Spy) FailReplication() {
	spy.isFailingReplication = true
}

func (spy *Spy) FailNextGet() {
	spy.isFailingGet = true
}
