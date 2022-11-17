package replication

import "time"

// SleepTimer implements the Timer interface providing a wrapper around basic thread sleep operations
type SleepTimer struct {
}

func NewSleepTimer() *SleepTimer {

	return &SleepTimer{}
}

// WaitMs sleeps the current routine for duration milliseconds
func (timer *SleepTimer) WaitMs(duration int) {

	var foo = time.Duration(duration)

	time.Sleep(foo * time.Millisecond)
}
