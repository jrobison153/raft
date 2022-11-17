package replication

const defaultPollPeriod = 50

// A Config provides fields that can be used to modify the way the replicator behaves
type Config struct {
	// JournalPollPeriod specifies the time in milliseconds to wait between checking the journal for
	// new journal entries that need to be replicated. Default value is 50ms
	JournalPollPeriod int
}

type Replicator interface {
	Start(Timer)
	TypeOfLogger() string
}

func NewDefaultConfig() *Config {

	return &Config{
		JournalPollPeriod: defaultPollPeriod,
	}
}
