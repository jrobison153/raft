# Unimplemented Test Scenarios

Just so I don't forget

## ClientAPI: Put Item

* Hash collision - unlikely to happen but needs to be considered

## ClientAPI: Get Item

* Hash collision - unlikely to happen but needs to be considered

## State machine

* Update state machine after log entry is committed. Early thoughts, create an interface and provide a default B-Tree implementation. Enhancements could include directions on how to create more implementations and configure which is loaded at runtime. Ideally these would be runtime dependencies not compile time but I'm not sure how to do that ...  yet (sockets or other IPC?)

## The Log

* Starting with just in memory need to consider memory cache and write to disk. 
* compaction