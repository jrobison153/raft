package replication

import (
	"github.com/jrobison153/raft/journal"
	"testing"
	"time"
)

const pollPeriod = 10

func TestWhenNoUncommittedJournalEntriesThenCommitNotCalledAfterReplication(t *testing.T) {

	journalSpy, _ := setup()

	// give the replicator time to replicate a little
	time.Sleep(5 * pollPeriod * time.Millisecond)

	if journalSpy.CommitCalled() {
		t.Errorf("Commit should not have been called as there was nothing to commit")
	}
}

func TestWhenUncommittedJournalEntriesThenTheyAreCommittedAfterReplication(t *testing.T) {

	journalSpy, _ := setup()

	appendResult := appendMany(journalSpy, 10)

	// give the replicator time to replicate a little
	time.Sleep(5 * pollPeriod * time.Millisecond)

	if !journalSpy.CommitCalledOnIndex(appendResult.Index) {
		t.Errorf("Commit should have been called on index %d", appendResult.Index)
	}
}

func TestWhenNoDataToReplicateThenReplicatorWaitsBetweenReplicationAttempts(t *testing.T) {

	_, timerSpy := setup()

	// give the replicator time to replicate a little
	time.Sleep(5 * pollPeriod * time.Millisecond)

	if timerSpy.WaitMsCalledWith() != pollPeriod {
		t.Errorf("Replication should have waited %dms between replications", pollPeriod)
	}
}

func TestWhenReplicatingDataThenReplicatorWaitsBetweenReplications(t *testing.T) {

	journalSpy, timerSpy := setup()

	appendMany(journalSpy, 5)

	// give the replicator time to replicate a little
	time.Sleep(5 * pollPeriod * time.Millisecond)

	if timerSpy.WaitMsCalledWith() != pollPeriod {
		t.Errorf("Replication should have waited %dms between replications", pollPeriod)
	}
}

func appendMany(journalSpy *journal.Spy, count int) journal.AppendResult {

	var appendResult journal.AppendResult

	for i := 0; i <= count; i++ {

		result := journalSpy.Append([]byte("some data"))

		appendResult = <-result
	}

	return appendResult
}

func setup() (*journal.Spy, *TimerSpy) {

	journalSpy := journal.NewJournalSpy()

	replicatorConfig := NewDefaultConfig()
	replicatorConfig.JournalPollPeriod = pollPeriod

	replicator := NewNoOpReplicator(journalSpy, replicatorConfig)

	timerSpy := NewTimerSpy()
	replicator.Start(timerSpy)

	return journalSpy, timerSpy
}
