package servicemanager

import (
	"sync"
	"time"

	"github.com/couchbase/cbauth/service"
	"github.com/couchbase/eventing/common"
	"github.com/couchbase/eventing/util"
)

const (
	metakvEventingPath       = "/eventing/"
	metakvAppsPath           = metakvEventingPath + "apps/"
	metakvAppSettingsPath    = metakvEventingPath + "appsettings/"     // function settings
	metakvConfigKeepNodes    = metakvEventingPath + "config/keepNodes" // Store list of eventing keepNodes
	metakvConfigPath         = metakvEventingPath + "config/settings"  // global settings
	metakvRebalanceTokenPath = metakvEventingPath + "rebalanceToken/"
	metakvRebalanceProgress  = metakvEventingPath + "rebalanceProgress/"
	metakvTempAppsPath       = metakvEventingPath + "tempApps/"
	metakvChecksumPath       = metakvEventingPath + "checksum/"
	metakvTempChecksumPath   = metakvEventingPath + "tempchecksum/"
	stopRebalance            = "stopRebalance"
)

const (
	rebalanceProgressUpdateTickInterval = time.Duration(3000) * time.Millisecond
)

const (
	// EventingPermissionManage for auditing
	EventingPermissionManage = "cluster.eventing.functions!manage"
)

const (
	headerKey                = "status"
	maxApplicationNameLength = 100
	maxAliasLength           = 20 // Technically, there isn't any limit on a JavaScript variable length.

	// Rebalance progress ticker ticks at every 3s, 1200 ticks would amount to 1hr.
	// If rebalance is stuck are specific progress % for longer than 1hr, rebalance
	// would be failed.
	rebalanceStalenessCounter = 1200
)

const (
	srcMapExt  = ".map.json"
	srcCodeExt = ".js"
)

const (
	maxHandlerSize = 128 * 1024
)

// ServiceMgr implements cbauth_service interface
type ServiceMgr struct {
	adminHTTPPort     string
	adminSSLPort      string
	auth              string
	certFile          string
	config            util.ConfigHolder
	ejectNodeUUIDs    []string
	eventingNodeAddrs []string
	failoverNotif     bool
	keepNodeUUIDs     []string
	keyFile           string
	mu                *sync.RWMutex
	uuid              string

	stopTracerCh chan struct{} // chan used to signal stopping of runtime.Trace

	nodeInfo         *service.NodeInfo
	rebalanceCtx     *rebalanceContext
	rebalancer       *rebalancer
	rebalanceRunning bool

	rebUpdateTicker *time.Ticker
	restPort        string
	servers         []service.NodeID
	state

	superSup common.EventingSuperSup
	waiters  waiters

	statusCodes   statusCodes
	statusPayload []byte
	errorCodes    map[int]errorPayload
}

type doneCallback func(err error, cancel <-chan struct{})
type progressCallback func(progress float64, cancel <-chan struct{})

type callbacks struct {
	done     doneCallback
	progress progressCallback
}

type rebalancer struct {
	cb     callbacks
	change service.TopologyChange

	c    chan struct{}
	done chan struct{}

	adminPort string
	keepNodes []string
}

type rebalanceContext struct {
	change service.TopologyChange
	rev    uint64
}

type waiter chan state
type waiters map[waiter]struct{}

type state struct {
	rebalanceID   string
	rebalanceTask *service.Task
	rev           uint64
	servers       []service.NodeID
}

type cleanup struct {
	canceled bool
	f        func()
}

type application struct {
	Name             string                 `json:"appname"`
	ID               int                    `json:"id"`
	DeploymentConfig depCfg                 `json:"depcfg"`
	AppHandlers      string                 `json:"appcode"`
	Settings         map[string]interface{} `json:"settings"`
}

type depCfg struct {
	Buckets        []bucket `json:"buckets"`
	MetadataBucket string   `json:"metadata_bucket"`
	SourceBucket   string   `json:"source_bucket"`
}

type bucket struct {
	Alias      string `json:"alias"`
	BucketName string `json:"bucket_name"`
}

type backlogStat struct {
	DcpBacklog uint64 `json:"dcp_backlog"`
}

type stats struct {
	DocTimerDebugStats              interface{} `json:"doc_timer_debug_stats,omitempty"`
	EventProcessingStats            interface{} `json:"event_processing_stats,omitempty"`
	EventsRemaining                 interface{} `json:"events_remaining,omitempty"`
	ExecutionStats                  interface{} `json:"execution_stats,omitempty"`
	FailureStats                    interface{} `json:"failure_stats,omitempty"`
	FunctionName                    interface{} `json:"function_name"`
	InternalVbDistributionStats     interface{} `json:"internal_vb_distribution_stats,omitempty"`
	LatencyStats                    interface{} `json:"latency_stats,omitempty"`
	LcbExceptionStats               interface{} `json:"lcb_exception_stats,omitempty"`
	PlannerStats                    interface{} `json:"planner_stats,omitempty"`
	PlasmaStats                     interface{} `json:"plasma_stats,omitempty"`
	SeqsProcessed                   interface{} `json:"seqs_processed,omitempty"`
	VbDcpEventsRemaining            interface{} `json:"dcp_event_backlog_per_vb,omitempty"`
	VbDistributionStatsFromMetadata interface{} `json:"vb_distribution_stats_from_metadata,omitempty"`
	WorkerPids                      interface{} `json:"worker_pids,omitempty"`
}

type config struct {
	RAMQuota       int    `json:"ram_quota"`
	MetadataBucket string `json:"metadata_bucket"`
}

type configResponse struct {
	Restart bool `json:"restart"`
}

type credsInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
