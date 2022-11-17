package log

import (
	"bytes"
	"errors"
	"testing"
)

var key []byte
var data []byte
var logger Logger

func setup() {

	key = []byte("12345")
	data = []byte("I am some sexy data!")

	logger = NewArrayLog()
}

func TestWhenDataAppendedToTheLogThenTheHeadEntryKeyIsSet(t *testing.T) {

	setup()

	logger.Append(key, data)

	headEntry, _ := logger.GetHead()

	if !bytes.Equal(key, headEntry.Key) {
		t.Errorf("The last appended item of the log should have been key '%s' but got '%s'",
			string(key),
			string(headEntry.Key))
	}
}

func TestWhenDataAppendedToTheLogThenTheHeadEntryDataIsSet(t *testing.T) {

	setup()

	logger.Append(key, data)

	headEntry, _ := logger.GetHead()

	if !bytes.Equal(data, headEntry.Data) {
		t.Errorf("The last appended item of the log should have been data '%s' but got '%s'",
			string(data),
			string(headEntry.Data))
	}
}

func TestGivenLogEmptyWhenGettingHeadThenAnErrorIsReturned(t *testing.T) {

	setup()

	_, appendErr := logger.GetHead()

	if !errors.Is(appendErr, ErrEmptyLog) {
		t.Errorf("Append should have returned an ErrEmptyLog, instead we got %v", appendErr)
	}
}
