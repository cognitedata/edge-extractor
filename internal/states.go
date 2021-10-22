package internal

import "time"

const (
	ProcessorStateRunning  = "RUNNING"
	ProcessorStateStarting = "STARTING"
	ProcessorStateShutdown = "SHUTDOWN"
	ProcessorStateStopped  = "STOPPED"
)

type ProcessorState struct {
	ID           uint64
	CurrentState string
	TargetState  string
}

type StateTracker struct {
	procStates []ProcessorState
}

func NewStateTracker() *StateTracker {
	return &StateTracker{}
}

func (intgr *StateTracker) SetProcessorTargetState(procId uint64, state string) {
	st := intgr.GetProcessorState(procId)
	if st == nil {
		intgr.procStates = append(intgr.procStates, ProcessorState{ID: procId, TargetState: state})
	} else {
		st.TargetState = state
	}
}

func (intgr *StateTracker) SetProcessorCurrentState(procId uint64, state string) {
	st := intgr.GetProcessorState(procId)
	if st == nil {
		intgr.procStates = append(intgr.procStates, ProcessorState{ID: procId, CurrentState: state})
	} else {
		st.CurrentState = state
	}
}

func (intgr *StateTracker) GetProcessorState(procId uint64) *ProcessorState {
	for i := range intgr.procStates {
		if intgr.procStates[i].ID == procId {
			return &intgr.procStates[i]
		}
	}
	return nil
}

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
