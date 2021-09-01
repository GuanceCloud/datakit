package nsq

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type stats struct {
	pts     []*io.Point
	tags    map[string]string
	lastErr error
	host    string
}

func newStats(serverhost string, tags map[string]string) *stats {
	return &stats{host: serverhost, tags: tags}
}

func (s *stats) parse(body []byte) ([]*io.Point, error) {
	var err error
	data := &NSQStatsData{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %s", err)
	}
	// Data was not parsed correctly attempt to use old format.
	if len(data.Version) < 1 {
		wrapper := &NSQStats{}
		err = json.Unmarshal(body, wrapper)
		if err != nil {
			return nil, fmt.Errorf("error parsing response: %s", err)
		}
		data = &wrapper.Data
	}

	s.dataStats(data)
	return s.result()
}

func (s *stats) result() ([]*io.Point, error) {
	return s.pts, s.lastErr
}

func (s *stats) dataStats(data *NSQStatsData) {
	tags := map[string]string{
		"server_host":    s.host,
		"server_version": data.Version,
	}
	for k, v := range s.tags {
		tags[k] = v
	}

	fields := make(map[string]interface{})
	if data.Health == "OK" {
		fields["server_count"] = int64(1)
	} else {
		fields["server_count"] = int64(0)
	}
	fields["topic_count"] = int64(len(data.Topics))

	if pt, err := io.MakePoint("nsq_server", tags, fields, time.Now()); err != nil {
		s.lastErr = err
	} else {
		s.pts = append(s.pts, pt)
	}

	for _, t := range data.Topics {
		s.topicStats(t, data.Version)
	}
}

func (s *stats) topicStats(t TopicStats, version string) {
	tags := map[string]string{
		"server_host":    s.host,
		"server_version": version,
		"topic":          t.Name,
	}
	for k, v := range s.tags {
		tags[k] = v
	}

	fields := map[string]interface{}{
		"depth":         t.Depth,
		"backend_depth": t.BackendDepth,
		"message_count": t.MessageCount,
		"channel_count": int64(len(t.Channels)),
	}

	if pt, err := io.MakePoint("nsq_topic", tags, fields, time.Now()); err != nil {
		s.lastErr = err
	} else {
		s.pts = append(s.pts, pt)
	}

	for _, c := range t.Channels {
		s.channelStats(c, version, t.Name)
	}
}

func (s *stats) channelStats(c ChannelStats, version, topic string) {
	tags := map[string]string{
		"server_host":    s.host,
		"server_version": version,
		"topic":          topic,
		"channel":        c.Name,
	}
	for k, v := range s.tags {
		tags[k] = v
	}

	fields := map[string]interface{}{
		"depth":           c.Depth,
		"backend_depth":   c.BackendDepth,
		"in_flight_count": c.InFlightCount,
		"deferred_count":  c.DeferredCount,
		"message_count":   c.MessageCount,
		"requeue_count":   c.RequeueCount,
		"timeout_count":   c.TimeoutCount,
		"client_count":    int64(len(c.Clients)),
	}

	if pt, err := io.MakePoint("nsq_channel", tags, fields, time.Now()); err != nil {
		s.lastErr = err
	} else {
		s.pts = append(s.pts, pt)
	}

	for _, cl := range c.Clients {
		s.clientStats(cl, version, topic, c.Name)
	}
}

func (s *stats) clientStats(c ClientStats, version, topic, channel string) {
	tags := map[string]string{
		"server_host":    s.host,
		"server_version": version,
		"topic":          topic,
		"channel":        channel,
		"client_id":      c.ID,
	}
	for k, v := range s.tags {
		tags[k] = v
	}

	if len(c.Name) > 0 {
		tags["client_name"] = c.Name
	}

	fields := map[string]interface{}{
		"client_hostname":   c.Hostname,
		"client_version":    c.Version,
		"client_address":    c.RemoteAddress,
		"client_user_agent": c.UserAgent,
		"ready_count":       c.ReadyCount,
		"in_flight_count":   c.InFlightCount,
		"message_count":     c.MessageCount,
		"finish_count":      c.FinishCount,
		"requeue_count":     c.RequeueCount,
		// TODO
		// "client_deflate":    strconv.FormatBool(c.Deflate),
		// "client_tls":        strconv.FormatBool(c.TLS),
		// "client_snappy":     strconv.FormatBool(c.Snappy),
	}

	if pt, err := io.MakePoint("nsq_client", tags, fields, time.Now()); err != nil {
		s.lastErr = err
	} else {
		s.pts = append(s.pts, pt)
	}
}

type NSQStats struct {
	Code int64        `json:"status_code"`
	Txt  string       `json:"status_txt"`
	Data NSQStatsData `json:"data"`
}

type NSQStatsData struct {
	Version   string       `json:"version"`
	Health    string       `json:"health"`
	StartTime int64        `json:"start_time"`
	Topics    []TopicStats `json:"topics"`
}

// e2e_processing_latency is not modeled
type TopicStats struct {
	Name         string         `json:"topic_name"`
	Depth        int64          `json:"depth"`
	BackendDepth int64          `json:"backend_depth"`
	MessageCount int64          `json:"message_count"`
	Paused       bool           `json:"paused"`
	Channels     []ChannelStats `json:"channels"`
}

// e2e_processing_latency is not modeled
type ChannelStats struct {
	Name          string        `json:"channel_name"`
	Depth         int64         `json:"depth"`
	BackendDepth  int64         `json:"backend_depth"`
	InFlightCount int64         `json:"in_flight_count"`
	DeferredCount int64         `json:"deferred_count"`
	MessageCount  int64         `json:"message_count"`
	RequeueCount  int64         `json:"requeue_count"`
	TimeoutCount  int64         `json:"timeout_count"`
	Paused        bool          `json:"paused"`
	Clients       []ClientStats `json:"clients"`
}

type ClientStats struct {
	Name                          string `json:"name"` // DEPRECATED 1.x+, still here as the structs are currently being shared for parsing v3.x and 1.x
	ID                            string `json:"client_id"`
	Hostname                      string `json:"hostname"`
	Version                       string `json:"version"`
	RemoteAddress                 string `json:"remote_address"`
	State                         int64  `json:"state"`
	ReadyCount                    int64  `json:"ready_count"`
	InFlightCount                 int64  `json:"in_flight_count"`
	MessageCount                  int64  `json:"message_count"`
	FinishCount                   int64  `json:"finish_count"`
	RequeueCount                  int64  `json:"requeue_count"`
	ConnectTime                   int64  `json:"connect_ts"`
	SampleRate                    int64  `json:"sample_rate"`
	Deflate                       bool   `json:"deflate"`
	Snappy                        bool   `json:"snappy"`
	UserAgent                     string `json:"user_agent"`
	TLS                           bool   `json:"tls"`
	TLSCipherSuite                string `json:"tls_cipher_suite"`
	TLSVersion                    string `json:"tls_version"`
	TLSNegotiatedProtocol         string `json:"tls_negotiated_protocol"`
	TLSNegotiatedProtocolIsMutual bool   `json:"tls_negotiated_protocol_is_mutual"`
}
