package replication

type TimerSpy struct {
	waitMsCallDuration int
}

func NewTimerSpy() *TimerSpy {

	return &TimerSpy{}
}

func (spy *TimerSpy) WaitMs(duration int) {

	spy.waitMsCallDuration = duration
}

func (spy *TimerSpy) WaitMsCalledWith() int {

	return spy.waitMsCallDuration
}
