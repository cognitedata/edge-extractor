package internal

import (
	"runtime/debug"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	cogClient                *CdfClient
	isStarted                bool
	stateTracker             *StateTracker
	globalCamPollingInterval time.Duration
	successCounter           uint64
	failureCounter           uint64
	extractorID              string
	ConfigObserver           *CdfConfigObserver // remote config observer
	enctyptionKey            string
}

func (p *Processor) SetEncryptionKey(key string) {
	p.enctyptionKey = key
}

func (intgr *Processor) Stop() {
	intgr.isStarted = false
}

func (intgr *Processor) stopProcessor(procId uint64) {
	log.Infof("Sending stop signal to processor %d ", procId)
	intgr.stateTracker.SetProcessorTargetState(procId, ProcessorStateStopped)
	if intgr.stateTracker.WaitForProcessorTargetState(procId, time.Second*120) {
		log.Infof("Processor %d has been stopped", procId)
	} else {
		log.Errorf("Failed to restart processor %d. Previous instance is still running", procId)
	}
}

func (intgr *Processor) reportRunStatus(camExternalID, status, msg string) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		log.Error(" Pipeliene monitoring failed to load configuration from CDF with error : ", stack)
	}
	exRun := core.CreateExtractionRun{ExternalID: intgr.extractorID, Status: status, Message: msg}
	intgr.cogClient.Client().ExtractionPipelines.CreateExtractionRuns(core.CreateExtractonRunsList{exRun})
}

// startSelfMonitoring run a status reporting look that periodically sends status reports to pipeline monitoring
func (intgr *Processor) startSelfMonitoring() {
	for {
		if intgr.successCounter > 0 && intgr.failureCounter == 0 {
			intgr.reportRunStatus("", core.ExtractionRunStatusSuccess, "all cameras operational")
		} else if intgr.successCounter > 0 && intgr.failureCounter > 0 {
			intgr.reportRunStatus("", core.ExtractionRunStatusSuccess, "some cameras not operational")
		} else {
			intgr.reportRunStatus("", core.ExtractionRunStatusSeen, "")
		}
		intgr.successCounter = 0
		intgr.failureCounter = 0
		time.Sleep(time.Second * 60)
	}
}
