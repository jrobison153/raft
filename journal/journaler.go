// Package journal contains interfaces and implementations of append only Item storage logs/journals
package journal

type Entry struct {
	Item []byte
}

type AppendResult struct {
	Index uint64
	Error error
}

type CommitResult struct {
	Error error
}

type AllUncommittedEntriesResult struct {

	// UncommittedEntries iterates over each Entry that has yet to be committed
	UncommittedEntries Iterator

	// HasUncommittedEntries will be true if there are uncommitted entries and false otherwise
	HasUncommittedEntries bool

	// CommitIndex is the index of the currently committed Entry in the journal
	CommitIndex int

	// HeadIndex is the index of the latest Entry in the journal
	HeadIndex int
}

// Iterator provides a read only view into the item backing a Journaler
type Iterator interface {
	Size() int
	Next() (Entry, error)
	HasNext() bool
}

type Journaler interface {
	Append(item []byte) chan AppendResult
	Commit(index uint64) chan CommitResult
	GetAllCommittedEntries() Iterator
	GetAllEntriesBetween(beginIndex uint64, endIndex uint64) (Iterator, error)
	GetAllUncommittedEntries() chan AllUncommittedEntriesResult
	GetHead() (Entry, error)
	NotifyOfAllCommitChanges(ch chan uint64)
	NotifyOfCommitOnIndexOnce(index uint64, ch chan bool) error
}
