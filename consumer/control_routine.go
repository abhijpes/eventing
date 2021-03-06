package consumer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/couchbase/eventing/logging"
	"github.com/couchbase/eventing/util"
	"github.com/couchbase/gocb"
)

func (c *Consumer) controlRoutine() {
	logPrefix := "Consumer::controlRoutine"

	for {
		select {
		case <-c.clusterStateChangeNotifCh:

			util.Retry(util.NewFixedBackoff(clusterOpRetryInterval), getEventingNodeAddrOpCallback, c)

			c.stopVbOwnerGiveupCh = make(chan struct{}, c.vbOwnershipGiveUpRoutineCount)
			c.stopVbOwnerTakeoverCh = make(chan struct{}, c.vbOwnershipTakeoverRoutineCount)

			logging.Infof("%s [%s:%s:%d] Got notification that cluster state has changed",
				logPrefix, c.workerName, c.tcpPort, c.Pid())

			c.vbsStreamClosedRWMutex.Lock()
			c.vbsStreamClosed = make(map[uint16]bool)
			c.vbsStreamClosedRWMutex.Unlock()

			c.isRebalanceOngoing = true
			logging.Infof("%s [%s:%s:%d] Updated isRebalanceOngoing to %v",
				logPrefix, c.workerName, c.tcpPort, c.Pid(), c.isRebalanceOngoing)
			go c.vbsStateUpdate()

		case <-c.signalSettingsChangeCh:

			logging.Infof("%s [%s:%s:%d] Got notification for settings change",
				logPrefix, c.workerName, c.tcpPort, c.Pid())

			settingsPath := metakvAppSettingsPath + c.app.AppName
			sData, err := util.MetakvGet(settingsPath)
			if err != nil {
				logging.Errorf("%s [%s:%s:%d] Failed to fetch updated settings from metakv, err: %v",
					logPrefix, c.workerName, c.tcpPort, c.Pid(), err)
				continue
			}

			settings := make(map[string]interface{})
			err = json.Unmarshal(sData, &settings)
			if err != nil {
				logging.Errorf("%s [%s:%s:%d] Failed to unmarshal settings received from metakv, err: %ru",
					logPrefix, c.workerName, c.tcpPort, c.Pid(), err)
				continue
			}

			if val, ok := settings["log_level"]; ok {
				c.logLevel = val.(string)
				logging.SetLogLevel(util.GetLogLevel(c.logLevel))
				c.sendLogLevel(c.logLevel, false)
			}

			if val, ok := settings["skip_timer_threshold"]; ok {
				c.skipTimerThreshold = int(val.(float64))
			}

			if val, ok := settings["vb_ownership_giveup_routine_count"]; ok {
				c.vbOwnershipGiveUpRoutineCount = int(val.(float64))
			}

			if val, ok := settings["vb_ownership_takeover_routine_count"]; ok {
				c.vbOwnershipTakeoverRoutineCount = int(val.(float64))
			}

		case <-c.restartVbDcpStreamTicker.C:

		retryVbsRemainingToRestream:
			c.Lock()
			vbsToRestream := make([]uint16, len(c.vbsRemainingToRestream))
			copy(vbsToRestream, c.vbsRemainingToRestream)
			c.Unlock()

			if len(vbsToRestream) == 0 {
				continue
			}

			// Verify if the app is deployed or not before trying to reopen vbucket DCP streams
			// for the ones which recently have returned STREAMEND. QE frequently does flush
			// on source bucket right after undeploy
			deployedApps := c.superSup.GetLocallyDeployedApps()
			if _, ok := deployedApps[c.app.AppName]; !ok {

				c.Lock()
				c.vbsRemainingToRestream = make([]uint16, 0)
				c.Unlock()

				logging.Infof("%s [%s:%s:%d] Discarding request to restream vbs: %v as the app has been undeployed",
					logPrefix, c.workerName, c.tcpPort, c.Pid(), util.Condense(vbsToRestream))
				continue
			}

			sort.Sort(util.Uint16Slice(vbsToRestream))
			logging.Infof("%s [%s:%s:%d] vbsToRestream len: %v dump: %v",
				logPrefix, c.workerName, c.tcpPort, c.Pid(), len(vbsToRestream), util.Condense(vbsToRestream))

			var vbsFailedToStartStream []uint16

			for _, vb := range vbsToRestream {
				if c.checkIfVbAlreadyOwnedByCurrConsumer(vb) {
					continue
				}

				// During Eventing+KV swap rebalance:
				// STREAMEND received because of outgoing KV node adds up entries in vbsToRestream,
				// but when eventing node receives rebalance notification it may not need to restream those
				// vbuckets as per the planner's output. Hence additional checking to verify if the worker
				// should own the vbucket stream
				if !c.checkIfCurrentConsumerShouldOwnVb(vb) {
					continue
				}

				var vbBlob vbucketKVBlob
				var cas gocb.Cas
				vbKey := fmt.Sprintf("%s::vb::%d", c.app.AppName, vb)

				logging.Infof("%s [%s:%s:%d] vb: %v, reclaiming it back by restarting dcp stream",
					logPrefix, c.workerName, c.tcpPort, c.Pid(), vb)
				util.Retry(util.NewFixedBackoff(bucketOpRetryInterval), getOpCallback, c, vbKey, &vbBlob, &cas, false)

				err := c.updateVbOwnerAndStartDCPStream(vbKey, vb, &vbBlob)
				if err != nil {
					c.vbsStreamRRWMutex.Lock()
					if _, ok := c.vbStreamRequested[vb]; ok {
						logging.Infof("%s [%s:%s:%d] vb: %d Purging entry from vbStreamRequested",
							logPrefix, c.workerName, c.tcpPort, c.Pid(), vb)

						delete(c.vbStreamRequested, vb)
					}
					c.vbsStreamRRWMutex.Unlock()

					vbsFailedToStartStream = append(vbsFailedToStartStream, vb)
				}
			}

			logging.Infof("%s [%s:%s:%d] vbsFailedToStartStream => len: %v dump: %v",
				logPrefix, c.workerName, c.tcpPort, c.Pid(), len(vbsFailedToStartStream), util.Condense(vbsFailedToStartStream))

			vbsToRestream = util.VbsSliceDiff(vbsFailedToStartStream, vbsToRestream)

			c.Lock()
			diff := util.VbsSliceDiff(vbsToRestream, c.vbsRemainingToRestream)
			c.vbsRemainingToRestream = diff
			vbsRemainingToRestream := len(c.vbsRemainingToRestream)
			c.Unlock()

			sort.Sort(util.Uint16Slice(diff))

			if vbsRemainingToRestream > 0 {
				logging.Infof("%s [%s:%s:%d] Retrying vbsToRestream, remaining len: %v dump: %v",
					logPrefix, c.workerName, c.tcpPort, c.Pid(), vbsRemainingToRestream, util.Condense(diff))
				goto retryVbsRemainingToRestream
			}

		case <-c.stopControlRoutineCh:
			logging.Infof("%s [%s:%s:%d] Exiting control routine",
				logPrefix, c.workerName, c.tcpPort, c.Pid())
			return
		}
	}
}
