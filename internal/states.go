package internal

import (
	"sync"
	"time"
)

const (
	ProcessorStateRunning  = "RUNNING"
	ProcessorStateStarting = "STARTING"
	ProcessorStateShutdown = "SHUTDOWN"
	ProcessorStateStopped  = "STOPPED"
	ProcessorStateNotFound = "NOT_FOUND"
)

type ProcessorState struct {
	ID           uint64
	CurrentState string
	TargetState  string
}

// StateTracker keep track of current and target states for all processors
type StateTracker struct {
	procStates []ProcessorState
	mux        *sync.RWMutex
}

func NewStateTracker() *StateTracker {
	return &StateTracker{mux: &sync.RWMutex{}}
}

func (intgr *StateTracker) SetProcessorTargetState(procId uint64, state string) {
	intgr.mux.Lock()
	defer intgr.mux.Unlock()
	st := intgr.getProcessorState(procId)
	if st.CurrentState == ProcessorStateNotFound {
		intgr.procStates = append(intgr.procStates, ProcessorState{ID: procId, TargetState: state})
	} else {
		st.TargetState = state
	}
}

func (intgr *StateTracker) SetProcessorCurrentState(procId uint64, state string) {
	intgr.mux.Lock()
	defer intgr.mux.Unlock()
	st := intgr.getProcessorState(procId)
	if st.CurrentState == ProcessorStateNotFound {
		intgr.procStates = append(intgr.procStates, ProcessorState{ID: procId, CurrentState: state})
	} else {
		st.CurrentState = state
	}
}

// getProcessorState returns process state , the method is for internal use only
func (intgr *StateTracker) getProcessorState(procId uint64) *ProcessorState {
	for i := range intgr.procStates {
		if intgr.procStates[i].ID == procId {
			return &intgr.procStates[i]
		}
	}
	return &ProcessorState{ID: procId, CurrentState: ProcessorStateNotFound, TargetState: ProcessorStateNotFound}
}

// GetProcessorState public version of getProcessorState
func (intgr *StateTracker) GetProcessorState(procId uint64) *ProcessorState {
	intgr.mux.RLock()
	defer intgr.mux.RUnlock()
	return intgr.getProcessorState(procId)
}

// WaitForProcessorTargetState blocks execution untill processor reaches target or wait operation times out .
func (intgr *StateTracker) WaitForProcessorTargetState(procId uint64, timeout time.Duration) bool {
	endTime := time.Now().Add(timeout)
	for {

		if time.Now().After(endTime) {
			return false
		}

		st := intgr.GetProcessorState(procId)
		if st.CurrentState == st.TargetState {
			return true
		}
		time.Sleep(time.Second * 1)
	}

}
