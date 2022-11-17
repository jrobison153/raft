package client

import (
	"errors"
	"strings"
)

type Spy struct {
	lastPutKey           string
	lastPutData          []byte
	isFailingPut         bool
	isFailingReplication bool
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

func (spy *Spy) TypeOfReplicator() string {
	//TODO implement me
	panic("implement me")
}

func (spy *Spy) Put(key string, data []byte) (chan bool, error) {

	var err error
	var replResult = true

	if spy.isFailingPut {

		err = errors.New("failing Put for test reasons")
	} else if spy.isFailingReplication {
		replResult = false
	} else {

		spy.lastPutKey = key
		spy.lastPutData = data
	}

	doneCh := make(chan bool)

	go func(ch chan bool, val bool) { ch <- replResult }(doneCh, replResult)

	return doneCh, err
}

func (spy *Spy) Get(key string) ([]byte, error) {

	var err error
	var data []byte

	if strings.Compare(key, spy.lastPutKey) != 0 {
		err = errors.New("clientPolicySpy: attempt to get item that was not previously Put")
	} else {
		data = spy.lastPutData
	}

	return data, err
}

func (spy *Spy) LastPutKey() string {

	return spy.lastPutKey
}

func (spy *Spy) LastDataKey() []byte {

	return spy.lastPutData
}

func (spy *Spy) FailNextPut() {
	spy.isFailingPut = true
}

func (spy *Spy) FailReplication() {
	spy.isFailingReplication = true
}
