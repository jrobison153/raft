// Package log contains interfaces and implementations of append only data storage logs/journals
package log

type Entry struct {
	Key  []byte
	Data []byte
}

// A Logger can append data to a data store
type Logger interface {
	Append(key []byte, data []byte) error
	GetHead() (Entry, error)
}
