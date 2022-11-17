package log

import "errors"

// ArrayLog implements an append only Last In First Out (LIFO) stack
type ArrayLog struct {
	theLog    []Entry
	headIndex int64
}

var (
	ErrEmptyLog = errors.New("attempt to get Head item from an empty log")
)

func NewArrayLog() *ArrayLog {

	return &ArrayLog{
		theLog:    make([]Entry, 0, 1024),
		headIndex: -1,
	}
}

// Append adds an Entry to the log as the new Head item with key and data set as the respective Entry
// data elements
func (logger *ArrayLog) Append(key []byte, data []byte) error {

	entry := Entry{
		Key:  key,
		Data: data,
	}

	logger.theLog = append(logger.theLog, entry)
	logger.headIndex += 1

	return nil
}

// GetHead returns the head Entry of the log, i.e. the last Entry that was added via Append.
// If the log is empty then an ErrEmptyLog is returned
func (logger *ArrayLog) GetHead() (Entry, error) {

	var headEntry Entry
	var err error

	if logger.headIndex >= 0 {

		headEntry = logger.theLog[logger.headIndex]
	} else {
		err = ErrEmptyLog
	}

	return headEntry, err
}
