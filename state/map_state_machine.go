package state

type MapStateMachine struct {
}

func NewMapStateMachine() *MapStateMachine {

	return &MapStateMachine{}
}

func (state MapStateMachine) GetValueForKey(key string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
