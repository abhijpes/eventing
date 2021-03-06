package supervisor

import (
	"fmt"

	"github.com/couchbase/eventing/common"
	"github.com/couchbase/eventing/logging"
)

// ClearEventStats flushes event processing stats
func (s *SuperSupervisor) ClearEventStats() {
	for _, p := range s.runningProducers {
		p.ClearEventStats()
	}
}

// DeployedAppList returns list of deployed lambdas running under super_supervisor
func (s *SuperSupervisor) DeployedAppList() []string {
	appList := make([]string, 0)

	for app := range s.runningProducers {
		appList = append(appList, app)
	}

	return appList
}

// GetEventProcessingStats returns dcp/timer event processing stats
func (s *SuperSupervisor) GetEventProcessingStats(appName string) map[string]uint64 {
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetEventProcessingStats()
	}
	return nil
}

// GetAppCode returns handler code for requested appname
func (s *SuperSupervisor) GetAppCode(appName string) string {
	logPrefix := "SuperSupervisor::GetAppCode"

	logging.Infof("%s [%d] Request for app: %v", logPrefix, len(s.runningProducers), appName)
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetAppCode()
	}
	return ""
}

// GetDebuggerURL returns the v8 debugger url for supplied appname
func (s *SuperSupervisor) GetDebuggerURL(appName string) string {
	logPrefix := "SuperSupervisor::GetDebuggerURL"

	logging.Infof("%s [%d] Request for app: %v", logPrefix, len(s.runningProducers), appName)
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetDebuggerURL()
	}
	return ""
}

// GetDeployedApps returns list of deployed apps and their last deployment time
func (s *SuperSupervisor) GetDeployedApps() map[string]string {
	s.appListRWMutex.RLock()
	defer s.appListRWMutex.RUnlock()

	deployedApps := make(map[string]string)
	for app, timeStamp := range s.deployedApps {
		deployedApps[app] = timeStamp
	}

	return deployedApps
}

// GetHandlerCode returns handler code for requested appname
func (s *SuperSupervisor) GetHandlerCode(appName string) string {
	logPrefix := "SuperSupervisor::GetHandlerCode"

	logging.Infof("%s [%d] Request for app: %v", logPrefix, len(s.runningProducers), appName)
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetHandlerCode()
	}
	return ""
}

// GetLatencyStats dumps stats from cpp world
func (s *SuperSupervisor) GetLatencyStats(appName string) map[string]uint64 {
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetLatencyStats()
	}
	return nil
}

// GetLocallyDeployedApps returns list of deployed apps and their last deployment time
func (s *SuperSupervisor) GetLocallyDeployedApps() map[string]string {
	s.appListRWMutex.RLock()
	defer s.appListRWMutex.RUnlock()

	locallyDeployedApps := make(map[string]string)
	for app, timeStamp := range s.locallyDeployedApps {
		locallyDeployedApps[app] = timeStamp
	}

	return locallyDeployedApps
}

// GetExecutionStats returns aggregated failure stats from Eventing.Producer instance
func (s *SuperSupervisor) GetExecutionStats(appName string) map[string]interface{} {
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetExecutionStats()
	}
	return nil
}

// GetFailureStats returns aggregated failure stats from Eventing.Producer instance
func (s *SuperSupervisor) GetFailureStats(appName string) map[string]interface{} {
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetFailureStats()
	}
	return nil
}

// GetLcbExceptionsStats returns libcouchbase exception stats from CPP workers
func (s *SuperSupervisor) GetLcbExceptionsStats(appName string) map[string]uint64 {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.GetLcbExceptionsStats()
	}
	return nil
}

// GetSeqsProcessed returns vbucket specific sequence nos processed so far
func (s *SuperSupervisor) GetSeqsProcessed(appName string) map[int]int64 {
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetSeqsProcessed()
	}
	return nil
}

// GetSourceMap returns source map for requested appname
func (s *SuperSupervisor) GetSourceMap(appName string) string {
	logPrefix := "SuperSupervisor::GetSourceMap"

	logging.Infof("%s [%d] Request for app: %v", logPrefix, len(s.runningProducers), appName)
	if p, ok := s.runningProducers[appName]; ok {
		return p.GetSourceMap()
	}
	return ""
}

// RestPort returns ns_server port(typically 8091/9000)
func (s *SuperSupervisor) RestPort() string {
	return s.restPort
}

// SignalStartDebugger kicks off V8 Debugger for a specific deployed lambda
func (s *SuperSupervisor) SignalStartDebugger(appName string) {
	logPrefix := "SuperSupervisor::SignalStartDebugger"

	p, ok := s.runningProducers[appName]
	if ok {
		p.SignalStartDebugger()
	} else {
		logging.Errorf("%s [%d] Request for app: %v didn't go through as Eventing.Producer instance isn't alive",
			logPrefix, len(s.runningProducers), appName)
	}
}

// SignalStopDebugger stops V8 Debugger for a specific deployed lambda
func (s *SuperSupervisor) SignalStopDebugger(appName string) {
	logPrefix := "SuperSupervisor::SignalStopDebugger"

	p, ok := s.runningProducers[appName]
	if ok {
		p.SignalStopDebugger()
	} else {
		logging.Errorf("%s [%d] Request for app: %v didn't go through as Eventing.Producer instance isn't alive",
			logPrefix, len(s.runningProducers), appName)
	}
}

// GetAppState returns current state of app
func (s *SuperSupervisor) GetAppState(appName string) int8 {
	switch s.appDeploymentStatus[appName] {
	case true:
		switch s.appProcessingStatus[appName] {
		case true:
			return common.AppStateEnabled
		case false:
			return common.AppStateDisabled
		}
	case false:
		switch s.appDeploymentStatus[appName] {
		case true:
			return common.AppStateUnexpected
		case false:
			return common.AppStateUndeployed
		}
	}
	return common.AppState
}

// GetDcpEventsRemainingToProcess returns remaining dcp events to process
func (s *SuperSupervisor) GetDcpEventsRemainingToProcess(appName string) uint64 {
	logPrefix := "SuperSupervisor::GetDcpEventsRemainingToProcess"

	p, ok := s.runningProducers[appName]
	if ok {
		return p.GetDcpEventsRemainingToProcess()
	}
	logging.Errorf("%s [%d] Request for app: %v didn't go through as Eventing.Producer instance isn't alive",
		logPrefix, len(s.runningProducers), appName)
	return 0
}

// VbDcpEventsRemainingToProcess returns remaining dcp events to process
func (s *SuperSupervisor) VbDcpEventsRemainingToProcess(appName string) map[int]int64 {
	logPrefix := "SuperSupervisor::VbDcpEventsRemainingToProcess"

	p, ok := s.runningProducers[appName]
	if ok {
		return p.VbDcpEventsRemainingToProcess()
	}
	logging.Errorf("%s [%d] Request for app: %v didn't go through as Eventing.Producer instance isn't alive",
		logPrefix, len(s.runningProducers), appName)
	return nil
}

// GetEventingConsumerPids returns map of Eventing.Consumer worker name and it's os pid
func (s *SuperSupervisor) GetEventingConsumerPids(appName string) map[string]int {
	logPrefix := "SuperSupervisor::GetEventingConsumerPids"

	p, ok := s.runningProducers[appName]
	if ok {
		return p.GetEventingConsumerPids()
	}
	logging.Errorf("%s [%d] Request for app: %v didn't go through as Eventing.Producer instance isn't alive",
		logPrefix, len(s.runningProducers), appName)
	return nil
}

// GetPlasmaStats returns internal stats from plasma
func (s *SuperSupervisor) GetPlasmaStats(appName string) (map[string]interface{}, error) {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.GetPlasmaStats()
	}

	return nil, fmt.Errorf("Eventing.Producer isn't alive")
}

// InternalVbDistributionStats returns internal state of vbucket ownership distribution on local eventing node
func (s *SuperSupervisor) InternalVbDistributionStats(appName string) map[string]string {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.InternalVbDistributionStats()
	}

	return nil
}

// VbDistributionStatsFromMetadata returns vbucket distribution across eventing nodes from metadata bucket
func (s *SuperSupervisor) VbDistributionStatsFromMetadata(appName string) map[string]map[string]string {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.VbDistributionStatsFromMetadata()
	}

	return nil
}

// PlannerStats returns vbucket distribution as per planner running on local eventing
// node for a given app
func (s *SuperSupervisor) PlannerStats(appName string) []*common.PlannerNodeVbMapping {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.PlannerStats()
	}

	return nil
}

// RebalanceTaskProgress reports vbuckets remaining to be transferred as per planner
// during the course of rebalance
func (s *SuperSupervisor) RebalanceTaskProgress(appName string) (*common.RebalanceProgress, error) {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.RebalanceTaskProgress(), nil
	}

	return nil, fmt.Errorf("Eventing.Producer isn't alive")
}

// TimerDebugStats captures timer related stats to assist in debugging mismtaches during rebalance
func (s *SuperSupervisor) TimerDebugStats(appName string) (map[int]map[string]interface{}, error) {
	p, ok := s.runningProducers[appName]
	if ok {
		return p.TimerDebugStats(), nil
	}

	return nil, fmt.Errorf("Eventing.Producer isn't alive")
}

// RebalanceStatus reports back status of rebalance for all running apps on current node
func (s *SuperSupervisor) RebalanceStatus() bool {
	logPrefix := "SuperSupervisor::RebalanceStatus"

	rebalanceStatuses := make(map[string]bool)
	for appName, p := range s.runningProducers {
		rebalanceStatuses[appName] = p.RebalanceStatus()
	}

	logging.Infof("%s [%d] Rebalance status from all running applications: %#v",
		logPrefix, len(s.runningProducers), rebalanceStatuses)

	for _, rebStatus := range rebalanceStatuses {
		if rebStatus {
			return rebStatus
		}
	}

	return false
}

// BootstrapAppList returns list of apps undergoing bootstrap
func (s *SuperSupervisor) BootstrapAppList() map[string]string {
	bootstrappingApps := make(map[string]string)

	s.appListRWMutex.RLock()
	defer s.appListRWMutex.RUnlock()

	for appName, ts := range s.bootstrappingApps {
		bootstrappingApps[appName] = ts
	}

	return bootstrappingApps
}
