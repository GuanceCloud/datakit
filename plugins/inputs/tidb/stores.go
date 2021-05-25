package tidb

import (
	"bytes"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type PDStores struct {
	Count  int `json:"count"`
	Stores []struct {
		Store struct {
			ID      int    `json:"id"`
			Address string `json:"address"`
			// Labels  []struct {
			// 	Key   string `json:"key"`
			// 	Value string `json:"value"`
			// } `json:"labels"`
			Version        string `json:"version"`
			GitHash        string `json:"git_hash"`
			DeployPath     string `json:"deploy_path"`
			StateName      string `json:"state_name"`
			PeerAddress    string `json:"peer_address"`
			StatusAddress  string `json:"status_address"`
			StartTimestamp int    `json:"start_timestamp"`
			// LastHeartbeat  int64  `json:"last_heartbeat"`
		} `json:"store,omitempty"`
		Status struct {
			Capacity        string    `json:"capacity"`
			Available       string    `json:"available"`
			UsedSize        string    `json:"used_size"`
			LeaderCount     int       `json:"leader_count"`
			LeaderWeight    int       `json:"leader_weight"`
			LeaderScore     int       `json:"leader_score"`
			LeaderSize      int       `json:"leader_size"`
			RegionCount     int       `json:"region_count"`
			RegionWeight    int       `json:"region_weight"`
			RegionScore     int       `json:"region_score"`
			RegionSize      int       `json:"region_size"`
			StartTs         time.Time `json:"start_ts"`
			LastHeartbeatTs time.Time `json:"last_heartbeat_ts"`
			Uptime          string    `json:"uptime"`
		} `json:"status"`
	} `json:"stores"`
}

func (p *PDStores) Metrics(measurement string, tags map[string]string, t time.Time) []byte {
	if p.Count == 0 {
		return nil
	}

	var buffer = bytes.Buffer{}
	for _, s := range p.Stores {
		_tags := make(map[string]string)
		_tags["id"] = strconv.Itoa(s.Store.ID)
		_tags["address"] = s.Store.Address
		_tags["version"] = s.Store.Version
		_tags["git_hash"] = s.Store.GitHash
		_tags["deploy_path"] = s.Store.DeployPath
		_tags["state_name"] = s.Store.StateName
		_tags["peer_address"] = s.Store.PeerAddress
		_tags["status_address"] = s.Store.StatusAddress
		for k, v := range tags {
			_tags[k] = v
		}

		fields := make(map[string]interface{})
		fields["capacity_KiB"] = toKiB(s.Status.Capacity)
		fields["available_KiB"] = toKiB(s.Status.Available)
		fields["used_size_KiB"] = toKiB(s.Status.UsedSize)
		fields["leader_count"] = s.Status.LeaderCount
		fields["leader_weight"] = s.Status.LeaderWeight
		fields["leader_score"] = s.Status.LeaderScore
		fields["leader_size"] = s.Status.LeaderSize
		fields["region_count"] = s.Status.RegionCount
		fields["region_weight"] = s.Status.RegionWeight
		fields["region_score"] = s.Status.RegionScore
		fields["region_size"] = s.Status.RegionSize
		fields["start_ts"] = s.Status.StartTs.UnixNano()
		fields["last_heartbeat_ts"] = s.Status.LastHeartbeatTs.UnixNano()
		uptime, _ := time.ParseDuration(s.Status.Uptime)
		fields["uptime_seconds"] = uptime.Seconds()

		data, err := io.MakeMetric(measurement, _tags, fields, t)
		if err != nil {
			l.Error(err)
			continue
		}
		buffer.Write(data)
		buffer.WriteString("\n")
	}
	return buffer.Bytes()
}
