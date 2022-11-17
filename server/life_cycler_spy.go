package server

type Spy struct {
	startPort uint32
}

func NewLifeCyclerSpy() *Spy {

	return &Spy{}
}

func (spy *Spy) Start(port uint32) {

	spy.startPort = port
}

func (spy *Spy) Stop() {
	//TODO implement me
	panic("implement me")
}

func (spy *Spy) StartCalledOnPort(port uint32) bool {

	return spy.startPort == port
}
