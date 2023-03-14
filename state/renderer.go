package state

type Renderer interface {
	ResolveRequestToData(request []byte) ([]byte, error)
	Start() error
	TypeOfLogger() string
}
