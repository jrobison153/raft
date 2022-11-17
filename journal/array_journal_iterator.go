package journal

import "errors"

type ArrayJournalIterator struct {
	data               []Entry
	currentIndex       int
	sizeOfBackingArray int
}

var (
	ErrNoMoreElements = errors.New("attempt to get next iterator element but iterator is at the end")
)

func NewArrayJournalIterator(backingArray []Entry) *ArrayJournalIterator {

	return &ArrayJournalIterator{
		data:               backingArray,
		currentIndex:       0,
		sizeOfBackingArray: len(backingArray),
	}
}

func (it *ArrayJournalIterator) Size() int {

	return it.sizeOfBackingArray
}

func (it *ArrayJournalIterator) Next() (Entry, error) {

	var entry Entry
	var err error

	if it.HasNext() {

		entry = it.data[it.currentIndex]
		it.currentIndex += 1
	} else {
		err = ErrNoMoreElements
	}

	return entry, err
}

func (it *ArrayJournalIterator) HasNext() bool {

	result := false

	if it.currentIndex < it.sizeOfBackingArray {
		result = true
	}

	return result
}
