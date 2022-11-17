package replicator

// NoOp is a simple implementation of the replicator interface that doesn't do anything, can simply
// be used to fulfill interface for Dependency Injection needs
type NoOp struct {
}

func NewNoOp() *NoOp {

	return &NoOp{}
}

func (repl *NoOp) Replicate(*Item, chan bool) {
}
