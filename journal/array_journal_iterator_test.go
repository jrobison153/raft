package journal

import (
	"reflect"
	"testing"
)

func TestGivenInitializedIteratorThenLengthOfUnderlyingArrayIsTheSizeOfTheIterator(t *testing.T) {

	backingArray := []Entry{{}, {}, {}, {}}
	arrayLen := len(backingArray)

	iterator := NewArrayJournalIterator(backingArray)

	iteratorSize := iterator.Size()

	if arrayLen != iteratorSize {
		t.Errorf("Underlying array length of %d was not equal to the size of the iterator %d",
			arrayLen,
			iteratorSize)
	}
}

func TestGivenEmptyIteratorThenHasNextReturnsFalse(t *testing.T) {

	iterator := NewArrayJournalIterator([]Entry{})

	hasNext := iterator.HasNext()

	if hasNext {
		t.Errorf("Iterator is empty, has next should have returned false")
	}
}

func TestGivenMoreElementsToIterateOverThenHasNextReturnsTrue(t *testing.T) {

	iterator := NewArrayJournalIterator([]Entry{{}, {}, {}})

	hasNext := iterator.HasNext()

	if !hasNext {
		t.Errorf("Iterator has more elements has next should have returned true")
	}
}

func TestWhenIteratorNotExhaustedThenHasNextReturnsTrue(t *testing.T) {

	iterator := NewArrayJournalIterator([]Entry{{}, {}, {}})

	iterator.Next()
	iterator.Next()

	hasNext := iterator.HasNext()

	if !hasNext {
		t.Errorf("Iterator has more elements but false was returned")
	}

}

func TestGivenEmptyBackingArrayWhenGettingNextAnErrorIsReturned(t *testing.T) {

	iterator := NewArrayJournalIterator([]Entry{})

	_, err := iterator.Next()

	if err != ErrNoMoreElements {
		t.Errorf("Iterator has no elements, next should have returned an error")
	}
}

func TestWhenIteratorExhaustedAndAttemptToGetNextThenAnErrorIsReturned(t *testing.T) {

	iterator := NewArrayJournalIterator([]Entry{{}, {}, {}})

	iterator.Next()
	iterator.Next()
	iterator.Next()
	_, err := iterator.Next()

	if err != ErrNoMoreElements {
		t.Errorf("Iterator has no elements, next should have returned an error")
	}
}

func TestWhenNextItemRetrievedThenTheCorrectEntryIsReturned(t *testing.T) {

	backingArray := []Entry{
		{
			Key:    []byte("a key"),
			RawKey: "a key",
			Data:   []byte("1"),
		},
	}

	iterator := NewArrayJournalIterator(backingArray)

	entry, _ := iterator.Next()

	if !reflect.DeepEqual(backingArray[0], entry) {
		t.Errorf("Entry should have been '%+v' but got '%+v'", backingArray[0], entry)
	}
}
