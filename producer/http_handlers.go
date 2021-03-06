package producer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	cm "github.com/couchbase/eventing/common"
	"github.com/couchbase/eventing/util"
)

// EventsProcessedPSec reports back aggregate of events processed/sec from all running consumers
func (p *Producer) EventsProcessedPSec(w http.ResponseWriter, r *http.Request) {
	var eventsProcessedPSec cm.EventProcessingStats

	for _, consumer := range p.runningConsumers {
		pStats := consumer.EventsProcessedPSec()

		eventsProcessedPSec.DcpEventsProcessedPSec += pStats.DcpEventsProcessedPSec
		eventsProcessedPSec.TimerEventsProcessedPSec += pStats.TimerEventsProcessedPSec
	}
	eventsProcessedPSec.Timestamp = time.Now().Format("2006-01-02T15:04:05Z")

	encodedStats, err := json.Marshal(&eventsProcessedPSec)
	if err != nil {
		fmt.Fprintf(w, "Failed to encode event processing stats")
		return
	}

	fmt.Fprintf(w, "%v", string(encodedStats))
}

// GetSettings dumps the event handler specific config
func (p *Producer) GetSettings(w http.ResponseWriter, r *http.Request) {
	encodedSettings, err := json.Marshal(p.app.Settings)
	if err != nil {
		fmt.Fprintf(w, "Failed to encode event handler settings")
		return
	}
	fmt.Fprintf(w, "%s", string(encodedSettings))
}

// UpdateSettings updates the event handler settings
func (p *Producer) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Failed to read contents from request body")
		return
	}

	appSettings := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &appSettings)
	if err != nil {
		fmt.Fprintf(w, "Failed to unmarshal contents from request body")
		return
	}

	err = util.MetakvSet(metakvAppSettingsPath+p.appName, reqBody, nil)
	if err != nil {
		fmt.Fprintf(w, "Failed to store new handler setting in metakv")
		return
	}

	fmt.Fprintf(w, "Successfully written new settings, app: %s will be reloaded with new config", p.appName)
}

// GetWorkerMap dumps the vbucket distribution across V8 workers
func (p *Producer) GetWorkerMap(w http.ResponseWriter, r *http.Request) {
	encodedWorkerMap, err := json.Marshal(p.workerVbucketMap)
	if err != nil {
		fmt.Fprintf(w, "Failed to encode worker vbucket map")
		return
	}
	fmt.Fprintf(w, "%s", string(encodedWorkerMap))
}

// GetNodeMap dumps vbucket distribution across eventing nodes
func (p *Producer) GetNodeMap(w http.ResponseWriter, r *http.Request) {
	encodedEventingMap, err := json.Marshal(p.vbEventingNodeAssignMap)
	if err != nil {
		fmt.Fprintf(w, "Failed to encode worker vbucket map")
		return
	}
	fmt.Fprintf(w, "%s", string(encodedEventingMap))
}

// GetConsumerVbProcessingStats dumps internal state of vbucket specific details, which is what's written to metadata bucket as well
func (p *Producer) GetConsumerVbProcessingStats(w http.ResponseWriter, r *http.Request) {
	vbStats := make(map[string]map[uint16]map[string]interface{}, 0)

	for _, consumer := range p.runningConsumers {
		consumerName := consumer.ConsumerName()
		stats := consumer.VbProcessingStats()
		vbStats[consumerName] = stats
	}

	encodedVbStats, err := json.Marshal(vbStats)
	if err != nil {
		fmt.Fprintf(w, "Failed to encode consumer vbstats")
		return
	}
	fmt.Fprintf(w, "%s", string(encodedVbStats))
}
