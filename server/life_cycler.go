package server

type LifeCycler interface {
	Start(port uint32)
	Stop()
}
