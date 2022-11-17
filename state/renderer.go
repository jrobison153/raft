package state

type Renderer interface {
	GetValueForKey(key string) ([]byte, error)
}
