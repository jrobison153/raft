package replicator

import "github.com/jrobison153/raft/replicator"

type Spy struct {
	replicatedItem  *replicator.Item
	replicateCalled bool
}

// Interface Functions

func (spy *Spy) Replicate(item *replicator.Item, doneCh chan bool) {

	spy.replicatedItem = item
	spy.replicateCalled = true

	doneCh <- true
}

// Spy Functions

func New() *Spy {

	return &Spy{}
}

func (spy *Spy) ReplicatedItem() *replicator.Item {

	return spy.replicatedItem
}

func (spy *Spy) ReplicateCalled() bool {

	return spy.replicateCalled
}
