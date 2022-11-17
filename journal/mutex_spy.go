package journal

import (
	"strings"
)

const lock = "Lock"
const unlock = "Unlock"

type MutexSpy struct {
	LockCalled        bool
	UnlockCalled      bool
	MethodCallJournal []string
}

func NewMutexSpy() *MutexSpy {

	spy := &MutexSpy{}
	spy.Reset()

	return spy
}

func (spy *MutexSpy) Lock() {
	spy.LockCalled = true
	spy.MethodCallJournal = append(spy.MethodCallJournal, lock)
}

func (spy *MutexSpy) Unlock() {
	spy.UnlockCalled = true
	spy.MethodCallJournal = append(spy.MethodCallJournal, unlock)
}

func (spy *MutexSpy) UnlockCalledBeforeLock() bool {

	lockCallPosition := findFirstPositionOf(spy.MethodCallJournal, lock)
	unlockCallPosition := findFirstPositionOf(spy.MethodCallJournal, unlock)

	wasLockCalledBeforeUnlock := false

	if lockCallPosition >= 0 && unlockCallPosition >= 0 {
		wasLockCalledBeforeUnlock = lockCallPosition < unlockCallPosition
	}

	return wasLockCalledBeforeUnlock
}

func findFirstPositionOf(items []string, targetName string) int {

	var targetIndex = -1

	for index, name := range items {

		if strings.Compare(targetName, name) == 0 {
			targetIndex = index
		}
	}

	return targetIndex
}

func (spy *MutexSpy) Reset() {

	spy.LockCalled = false
	spy.UnlockCalled = false
	spy.MethodCallJournal = make([]string, 0, 16)
}
