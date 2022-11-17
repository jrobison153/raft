package replication

type Timer interface {
	WaitMs(int)
}
