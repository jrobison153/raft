package replication

import (
	"testing"
	"time"
)

func TestWhenWaitingMsThenTheCurrentRoutineSleeps(t *testing.T) {

	timer := NewSleepTimer()

	startTime := time.Now().UnixMilli()
	timer.WaitMs(50)
	stopTime := time.Now().UnixMilli()

	routineDelay := stopTime - startTime

	if routineDelay <= 35 || routineDelay >= 60 {
		t.Errorf("Expected routine to sleep roughly 50ms")
	}
}
