package integrations

import (
	"runtime/debug"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)

// BaseIntegration is a base class for all integrations. Integrations are long running processes that internally run one or more processors (goroutines)
// All processors share that same logic but configured differently. StateTracker is used to track the state of all processors and to control them.
type BaseIntegration struct {
	ID                  string
	CogClient           *internal.CdfClient
	IsRunning           bool
	StateTracker        *internal.StateTracker
	extractorID         string
	ConfigObserver      *internal.CdfConfigObserver // remote config observer
	disableRunReporting bool
}

func NewIntegration(id string, cogClient *internal.CdfClient, extractorID string, configObserver *internal.CdfConfigObserver) *BaseIntegration {
	return &BaseIntegration{ID: id,
		CogClient:      cogClient,
		extractorID:    extractorID,
		ConfigObserver: configObserver,
		StateTracker:   internal.NewStateTracker(),
	}
}

func (intgr *BaseIntegration) Stop() {
	intgr.IsRunning = false
}

func (intgr *BaseIntegration) DisableRunReporting(state bool) {
	intgr.disableRunReporting = state
}

func (intgr *BaseIntegration) StopProcessor(procId uint64) {
	procState := intgr.StateTracker.GetProcessorState(procId)
	if procState.CurrentState == internal.ProcessorStateStopped || procState.CurrentState == internal.ProcessorStateShutdown || procState.CurrentState == internal.ProcessorStateNotFound {
		log.Info("Processor is already stopped or not found")
		return
	}
	log.Infof("Sending stop signal to processor %d ", procId)
	intgr.StateTracker.SetProcessorTargetState(procId, internal.ProcessorStateStopped)
	if intgr.StateTracker.WaitForProcessorTargetState(procId, time.Second*120) {
		log.Infof("Processor %d has been stopped", procId)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", procId)
	}
}

func (intgr *BaseIntegration) ReportRunStatus(camExternalID, status, msg string) {
	if intgr.disableRunReporting {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error(" Pipeliene monitoring failed to report run status due to the error : ", stack)
		}
	}()
	exRun := core.CreateExtractionRun{ExternalID: intgr.extractorID, Status: status, Message: msg}
	client := intgr.CogClient.Client()
	if client == nil {
		log.Error("Cdf client is not initialized")
		return
	}
	client.ExtractionPipelines.CreateExtractionRuns(core.CreateExtractonRunsList{exRun})
}
