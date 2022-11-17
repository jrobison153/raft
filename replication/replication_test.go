package replication

import "testing"

func TestWhenDefaultConfigCreatedThenTheJournalPollPeriodValueIsSet(t *testing.T) {

	config := NewDefaultConfig()

	if config.JournalPollPeriod != defaultPollPeriod {
		t.Errorf("JournalPollPeriod should have been defaulted to %d but was %d",
			defaultPollPeriod,
			config.JournalPollPeriod)
	}
}