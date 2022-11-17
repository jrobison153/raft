package replicator

type Quorum struct {
}

func NewQuorum() *Quorum {

	return &Quorum{}
}

func (quorum Quorum) Replicate(item *Item, doneCh chan bool) {

	// TODO implement this
	doneCh <- true
}
