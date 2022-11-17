// Package log contains interfaces and implementations of append only data storage logs/journals
package journal

type Entry struct {
	Key  []byte
	Data []byte
}

type Journaler interface {
	Append(key []byte, data []byte) (uint64, error)
	Commit(index uint64) error
	GetHead() (Entry, error)
	NotifyOfCommitOnIndexOnce(index uint64, ch chan bool) error
}
