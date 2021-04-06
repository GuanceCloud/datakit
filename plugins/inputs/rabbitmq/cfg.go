package rabbitmq

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"time"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"net/http"
	"fmt"
	"encoding/json"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = `rabbitmq`
	l         = logger.DefaultSLogger(inputName)
	collectCache []inputs.Measurement
	sample    = `
[[inputs.rabbitmq]]
	# rabbitmq url ,required
	url = "http://localhost:15672"

	# rabbitmq user, required
	username = "guest"    

	# rabbitmq password, required
  	password = "guest"

	# ##(optional) collection interval, default is 30s
	# interval = "30s"
  
	## Optional TLS Config
	# tls_ca = "/xxx/ca.pem"
	# tls_cert = "/xxx/cert.cer"
	# tls_key = "/xxx/key.key"
	## Use TLS but skip chain & host verification
	insecure_skip_verify = false

`
)

type Input struct {
	Url      string           `toml:"url"`
	Username string           `toml:"username"`
	Password string           `toml:"password"`
	Interval datakit.Duration `toml:"interval"`
	tls.ClientConfig

	// HTTP client
	client *http.Client

	start time.Time
}

type OverviewResponse struct {
	Version      string        `json:"rabbitmq_version"`
	ClusterName  string        `json:"cluster_name"`
	MessageStats *MessageStats `json:"message_stats"`
	ObjectTotals *ObjectTotals `json:"object_totals"`
	QueueTotals  *QueueTotals  `json:"queue_totals"`
	Listeners    []Listeners   `json:"listeners"`
}

type Listeners struct {
	Protocol string `json:"protocol"`
}

// Details ...
type Details struct {
	Rate float64 `json:"rate"`
}

// MessageStats ...
type MessageStats struct {
	Ack                     int64
	AckDetails              Details `json:"ack_details"`
	Deliver                 int64
	DeliverDetails          Details `json:"deliver_details"`
	DeliverGet              int64   `json:"deliver_get"`
	DeliverGetDetails       Details `json:"deliver_get_details"`
	Publish                 int64
	PublishDetails          Details `json:"publish_details"`
	Redeliver               int64
	RedeliverDetails        Details `json:"redeliver_details"`
	PublishIn               int64   `json:"publish_in"`
	PublishInDetails        Details `json:"publish_in_details"`
	PublishOut              int64   `json:"publish_out"`
	PublishOutDetails       Details `json:"publish_out_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails Details `json:"return_unroutable_details"`
}

// ObjectTotals ...
type ObjectTotals struct {
	Channels    int64
	Connections int64
	Consumers   int64
	Exchanges   int64
	Queues      int64
}

type QueueTotals struct {
	Messages                   int64
	MessagesDetail             Details `json:"messages_details"`

	MessagesReady              int64 `json:"messages_ready"`
	MessagesReadyDetail        Details `json:"messages_ready_details"`

	MessagesUnacknowledged     int64 `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetail     Details `json:"messages_unacknowledged_details"`

}


type Exchange struct {
	Name         string
	MessageStats `json:"message_stats"`
	Type         string
	Internal     bool
	Vhost        string
	Durable      bool
	AutoDelete   bool `json:"auto_delete"`
}











func (n *Input) createHttpClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Second * 20,
	}

	return client, nil
}

func (n *Input) requestJSON(u string, target interface{}) error {

	u = fmt.Sprintf("%s%s", n.Url, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(n.Username, n.Password)

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(target)

	return nil
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}


func newRateFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}