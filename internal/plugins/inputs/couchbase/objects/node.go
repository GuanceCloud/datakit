// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	// System Stats Keys.
	CPUUtilizationRate = "cpu_utilization_rate"
	SwapTotal          = "swap_total"
	SwapUsed           = "swap_used"
	MemTotal           = "mem_total"
	MemFree            = "mem_free"

	// Interesting Stats Keys.
	CmdGet                                 = "cmd_get"
	CouchDocsActualDiskSize                = "couch_docs_actual_disk_size"
	CouchDocsDataSize                      = "couch_docs_data_size"
	CouchSpatialDataSize                   = "couch_spatial_data_size"
	CouchSpatialDiskSize                   = "couch_spatial_disk_size"
	CouchViewsActualDiskSize               = "couch_views_actual_disk_size"
	CouchViewsDataSize                     = "couch_views_data_size"
	CurrItems                              = "curr_items"
	CurrItemsTot                           = "curr_items_tot"
	EpBgFetched                            = "ep_bg_fetched"
	GetHits                                = "get_hits"
	InterestingStatsMemUsed                = "mem_used"
	Ops                                    = "ops"
	InterestingStatsVbActiveNumNonResident = "vb_active_num_non_resident"
	VbReplicaCurrItems                     = "vb_replica_curr_items"

	// Counters Keys.
	RebalanceSuccess        = "rebalance_success"
	RebalanceStart          = "rebalance_start"
	RebalanceFail           = "rebalance_fail"
	RebalanceStop           = "rebalance_stop"
	FailoverNode            = "failover_node"
	Failover                = "failover"
	FailoverComplete        = "failover_complete"
	FailoverIncomplete      = "failover_incomplete"
	GracefulFailoverStart   = "graceful_failover_start"
	GracefulFailoverSuccess = "graceful_failover_success"
	GracefulFailoverFail    = "graceful_failover_fail"
)

type Nodes struct {
	Name                   string            `json:"name"`
	Nodes                  []Node            `json:"nodes"`
	Buckets                map[string]string `json:"buckets"`        //
	RemoteClusters         map[string]string `json:"remoteClusters"` //
	Alerts                 []interface{}     `json:"alerts"`
	AlertsSilenceURL       string
	RebalanceStatus        string                 `json:"rebalanceStatus"`
	RebalanceProgressURI   string                 `json:"rebalanceProgressUri"` //
	StopRebalanceURI       string                 `json:"stopRebalanceUri"`     //
	NodeStatusesURI        string                 `json:"nodeStatusesUri"`      //
	MaxBucketCount         int                    `json:"maxBucketCount"`
	AutoCompactionSettings map[string]interface{} `json:"autoCompactionSettings"` //
	Tasks                  map[string]string      `json:"tasks"`                  //
	Counters               map[string]float64     `json:"counters"`
	IndexStatusURI         string                 `json:"indexStatusURI"`      //
	CheckPermissionsURI    string                 `json:"checkPermissionsURI"` //
	ServerGroupsURI        string                 `json:"serverGroupsUri"`     //
	ClusterName            string                 `json:"clusterName"`
	Balanced               bool                   `json:"balanced"`
	MemoryQuota            int                    `json:"memoryQuota"`
	IndexMemoryQuota       int                    `json:"indexMemoryQuota"`
	FtsMemoryQuota         int                    `json:"ftsMemoryQuota"`
	CbasMemoryQuota        int                    `json:"cbasMemoryQuota"`
	EventingMemoryQuota    int                    `json:"eventingMemoryQuota"`
	StorageTotals          StorageTotals          `json:"storageTotals"`
}

type Links struct {
	Tasks struct {
		URI string `json:"uri"`
	} `json:"tasks"`
	Buckets struct {
		URI                       string `json:"uri"`
		TerseBucketsBase          string `json:"terseBucketsBase"`
		TerseStreamingBucketsBase string `json:"terseStreamingBucketsBase"`
	} `json:"buckets"`
	RemoteClusters struct {
		URI         string `json:"uri"`
		ValidateURI string `json:"validateURI"`
	} `json:"remoteClusters"`
}

type StorageTotals struct {
	RAM struct {
		Total             float64 `json:"total"`
		QuotaTotal        float64 `json:"quotaTotal"`
		QuotaUsed         float64 `json:"quotaUsed"`
		Used              float64 `json:"used"`
		UsedByData        float64 `json:"usedByData"`
		QuotaUsedPerNode  float64 `json:"quotaUsedPerNode"`
		QuotaTotalPerNode float64 `json:"quotaTotalPerNode"`
	} `json:"ram"`
	Hdd struct {
		Total      float64 `json:"total"`
		QuotaTotal float64 `json:"quotaTotal"`
		Used       float64 `json:"used"`
		UsedByData float64 `json:"usedByData"`
		Free       float64 `json:"free"`
	} `json:"hdd"`
}

// Node struct itself
// contains a lot more fields when listed in a bucketinfo struct
// @ /pools/default/buckets

type Node struct {
	SystemStats          map[string]float64          `json:"systemStats,omitempty"`
	InterestingStats     map[string]float64          `json:"interestingStats"`
	Uptime               string                      `json:"uptime"`
	MemoryTotal          float64                     `json:"memoryTotal"`
	MemoryFree           float64                     `json:"memoryFree"`
	CouchAPIBaseHTTPS    string                      `json:"couchApiBaseHTTPS,omitempty"`
	CouchAPIBase         string                      `json:"couchApiBase"`
	McdMemoryReserved    float64                     `json:"mcdMemoryReserved"`
	McdMemoryAllocated   float64                     `json:"mcdMemoryAllocated"`
	Replication          float64                     `json:"replication,omitempty"`
	ClusterMembership    string                      `json:"clusterMembership"`
	RecoveryType         string                      `json:"recoveryType"`
	Status               string                      `json:"status"`
	OtpNode              string                      `json:"otpNode"`
	ThisNode             bool                        `json:"thisNode,omitempty"`
	OtpCookie            string                      `json:"otpCookie,omitempty"`
	Hostname             string                      `json:"hostname"`
	ClusterCompatibility int                         `json:"clusterCompatibility"`
	Version              string                      `json:"version"`
	Os                   string                      `json:"os"`
	CPUCount             interface{}                 `json:"cpuCount,omitempty"`
	Ports                *Ports                      `json:"ports,omitempty"`
	Services             []string                    `json:"services,omitempty"`
	AlternateAddresses   *AlternateAddressesExternal `json:"alternateAddresses,omitempty"`
}

type Ports struct {
	HTTPSMgmt int `json:"httpsMgmt"`
	HTTPSCAPI int `json:"httpsCAPI"`
	Proxy     int `json:"proxy"`
	Direct    int `json:"direct"`
}

type AlternateAddresses struct {
	External *AlternateAddressesExternal `json:"external,omitempty"`
}

// AlternateAddressesExternal AlternateAddresses defines a K8S node address and port mapping for
// use by clients outside of the pod network.  Hostname must be set,
// ports are ignored if zero.
type AlternateAddressesExternal struct {
	// Hostname is the host name to connect to (typically a L3 address)
	Hostname string `url:"hostname" json:"hostname"`
	// Ports is the map of service to external ports
	Ports *AlternateAddressesExternalPorts `url:"" json:"ports,omitempty"`
}

type AlternateAddressesExternalPorts struct {
	// AdminPort is the admin service K8S node port (mapped to 8091)
	AdminServicePort int32 `url:"mgmt,omitempty" json:"mgmt"`
	// AdminPortSSL is the admin service K8S node port (mapped to 18091)
	AdminServicePortTLS int32 `url:"mgmtSSL,omitempty" json:"mgmtSSL"`
	// ViewServicePort is the view service K8S node port (mapped to 8092)
	ViewServicePort int32 `url:"capi,omitempty" json:"capi"`
	// ViewServicePortSSL is the view service K8S node port (mapped to 8092)
	ViewServicePortTLS int32 `url:"capiSSL,omitempty" json:"capiSSL"`
	// QueryServicePort is the query service K8S node port (mapped to 8093)
	QueryServicePort int32 `url:"n1ql,omitempty" json:"n1ql"`
	// QueryServicePortTLS is the query service K8S node port (mapped to 18093)
	QueryServicePortTLS int32 `url:"n1qlSSL,omitempty" json:"n1qlSSL"`
	// FtsServicePort is the full text search service K8S node port (mapped to 8094)
	FtsServicePort int32 `url:"fts,omitempty" json:"fts"`
	// FtsServicePortTLS is the full text search service K8S node port (mapped to 18094)
	FtsServicePortTLS int32 `url:"ftsSSL,omitempty" json:"ftsSSL"`
	// AnalyticsServicePort is the analytics service K8S node port (mapped to 8095)
	AnalyticsServicePort int32 `url:"cbas,omitempty" json:"cbas"`
	// AnalyticsServicePortTLS is the analytics service K8S node port (mapped to 18095)
	AnalyticsServicePortTLS int32 `url:"cbasSSL,omitempty" json:"cbasSSL"`
	// DataServicePort is the data service K8S node port (mapped to 11210)
	DataServicePort int32 `url:"kv,omitempty" json:"kv"`
	// DataServicePortSSL is the data service K8S node port (mapped to 11207)
	DataServicePortTLS int32 `url:"kvSSL,omitempty" json:"kvSSL"`
}
