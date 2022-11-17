package replicator

type Spy struct {
	replicatedItem  *Item
	replicateCalled bool
}

// Interface Functions

func (spy *Spy) Replicate(item *Item, doneCh chan bool) {

	spy.replicatedItem = item
	spy.replicateCalled = true

	doneCh <- true
}

// Spy Functions

func NewSpy() *Spy {

	return &Spy{}
}

func (spy *Spy) ReplicatedItem() *Item {

	return spy.replicatedItem
}

func (spy *Spy) ReplicateCalled() bool {

	return spy.replicateCalled
}
