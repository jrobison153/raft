// Package replicator contains interface, types, and implementations that can be used to replicate
// data
package replicator

// An Item contains data that can be replicated
type Item struct {
	Key  []byte
	Data []byte
}

// A Replicator can replicate data
type Replicator interface {
	Replicate(item *Item, doneCh chan bool)
}
