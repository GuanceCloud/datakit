package redis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"github.com/go-redis/redis"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l         *logger.Logger
	inputName = "redis"
)

func (_ *Redis) Catalog() string {
	return "db"
}

func (_ *Redis) SampleConfig() string {
	return configSample
}

func (_ *Redis) Description() string {
	return ""
}

func (_ *Redis) Gather() error {
	return nil
}

func (_ *Redis) Init() error {
	return nil
}

func (r *Redis) Run() {
	l = logger.SLogger("redis")

	l.Info("redis input started...")

	r.IntervalDuration = 10 * time.Minute

	if r.Interval != "" {
		du, err := time.ParseDuration(r.IntervalDuration)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", r.Interval, err.Error())
		} else {
			r.IntervalDuration = du
		}
	}

	// 指标集名称
	if r.MetricName == "" {
		r.MetricName = inputName
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// todo
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (r *Redis) collectMetrics() error {
	err := r.collectInfoMetrics()
	if err != nil {
		return err
	}

	r.collectKeysLength()

	err = r.collectSlowlog()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) collectInfoMetrics() error {
	info, err := r.client.Info("ALL").Result()
	if err != nil {
		return err
	}

	rdr := strings.NewReader(info)

	fmt.Println("======>", rdr)
}

// func (r *Redis) collectInfoMetrics() error {
// 	start := time.Now()
// 	info, err := redis.String(r.db.Do("INFO", "ALL"))
// 	if err != nil {
// 		l.Errorf("Failed to run info command. %s", err.Error())
// 		return err
// 	}
// 	elapsed := time.Since(start)
// 	latencyMs := util.Round(float64(elapsed)/float64(time.Millisecond), 2)

// 	fields := map[string]interface{}

// 	fields["latency_ms"] = latencyMs

// 	lines := strings.Split(info, "\r\n")
// 	// Compatible with the raw string literals for test reason.
// 	if len(lines) == 1 {
// 		lines = strings.Split(info, "\n")
// 	}

// 	for _, line := range lines {
// 		if line == "" {
// 			continue
// 		}
// 		if re, _ := regexp.MatchString("^#", line); re {
// 			continue
// 		}

// 		record := strings.SplitN(line, ":", 2)
// 		if len(record) < 2 {
// 			continue
// 		}
// 		key, value := record[0], record[1]

// 		if re, _ := regexp.MatchString(`^db\d+`, key); re {
// 			r.collectDBMetrics(key, value)
// 			continue
// 		}

// 		val, err := strconv.ParseFloat(value, 64)
// 		if err != nil {
// 			continue
// 		}

// 		if name, ok := GAUGES[key]; ok {
// 			fields[name] = val
// 		}

// 		if name, ok := RATES[key]; ok {
// 			fields[name] = val
// 		}
// 	}

// 	r.collectReplicaMetrics(lines)
// 	return nil
// }

// func (r *Redis) collectDBMetrics(key, value string) {
// 	kv := strings.SplitN(value, ",", 3)
// 	if len(kv) != 3 {
// 		return
// 	}
// 	keys, expired := kv[0], kv[1]

// 	totalKeys, err := extractVal(keys)
// 	if err != nil {
// 		log.Warnf("Failed to parse db keys. %s", err)
// 	}

// 	expiredKeys, err := extractVal(expired)
// 	if err != nil {
// 		log.Warnf("Failed to parse db expired. %s", err)
// 	}

// 	dbTags := append(tags, "redis_db:"+key)
// 	persistKeys := totalKeys - expiredKeys
// 	fields := map[string]interface{}{
// 		"persist":         persistKeys,
// 		"persist.percent": 100 * persistKeys / totalKeys,
// 		"expires.percent": 100 * expiredKeys / totalKeys,
// 	}
// }

// func (r *Redis) collectReplicaMetrics(lines) {
// 	var masterDownSeconds, masterOffset, slaveOffset float64
// 	var masterStatus, slaveID, ip, port string
// 	var err error
// 	for _, line := range lines {
// 		record := strings.SplitN(line, ":", 2)
// 		if len(record) < 2 {
// 			continue
// 		}
// 		key, value := record[0], record[1]

// 		if key == "master_repl_offset" {
// 			masterOffset, _ = strconv.ParseFloat(value, 64)
// 		}

// 		if key == "master_link_down_since_seconds" {
// 			masterDownSeconds, _ = strconv.ParseFloat(value, 64)
// 		}

// 		if key == "master_link_status" {
// 			masterStatus = value
// 		}

// 		if re, _ := regexp.MatchString(`^slave\d+`, key); re {
// 			slaveID = strings.TrimPrefix(key, "slave")
// 			kv := strings.SplitN(value, ",", 5)
// 			if len(kv) != 5 {
// 				continue
// 			}

// 			split := strings.Split(kv[0], "=")
// 			if len(split) != 2 {
// 				log.Warnf("Failed to parse slave ip. %s", err)
// 				continue
// 			}
// 			ip = split[1]

// 			split = strings.Split(kv[1], "=")
// 			if err != nil {
// 				log.Warnf("Failed to parse slave port. %s", err)
// 				continue
// 			}
// 			port = split[1]

// 			split = strings.Split(kv[3], "=")
// 			if err != nil {
// 				log.Warnf("Failed to parse slave offset. %s", err)
// 				continue
// 			}
// 			slaveOffset, _ = strconv.ParseFloat(split[1], 64)
// 		}
// 	}

// 	delay := masterOffset - slaveOffset
// 	slaveTags := append(tags, "slave_ip:"+ip, "slave_port:"+port, "slave_id:"+slaveID)
// 	if delay >= 0 {
// 		fields["replication_delay"] = delay
// 	}

// 	if masterStatus != "" {
// 		fields["master_link_down_since_seconds"] = masterDownSeconds
// 	}
// }

// func (r *Redis) collectKeysLength() {
// 	for _, key := range r.Keys {
// 		found := false
// 		keyTags := append(tags, "key:"+key)

// 		for _, op := range []string{
// 			"HLEN",
// 			"LLEN",
// 			"SCARD",
// 			"ZCARD",
// 			"PFCOUNT",
// 			"STRLEN",
// 		} {
// 			if val, err := c.Do(op, key); err == nil && val != nil {
// 				found = true
// 				fields["length"] = val
// 				break
// 			}
// 		}

// 		if !found {
// 			if r.WarnOnMissingKeys {
// 				log.Warnf("%s key not found in redis", key)
// 			}

// 			fields["length"] = 0
// 		}
// 	}
// }

// func (r *Redis) collectSlowlog() error {
// 	var maxSlowEntries, defaultMaxSlowEntries float64
// 	defaultMaxSlowEntries = 128
// 	if r.SlowlogMaxLen > 0 {
// 		maxSlowEntries = r.SlowlogMaxLen
// 	} else {
// 		if config, err := redis.Strings(c.Do("CONFIG", "GET", "slowlog-max-len")); err == nil {
// 			fields, err := extractConfig(config)
// 			if err != nil {
// 				return nil
// 			}

// 			maxSlowEntries = fields["slowlog-max-len"].(float64)
// 			if maxSlowEntries > defaultMaxSlowEntries {
// 				maxSlowEntries = defaultMaxSlowEntries
// 			}
// 		} else {
// 			maxSlowEntries = defaultMaxSlowEntries
// 		}
// 	}

// 	// Generate a unique id for this instance to be persisted across runs
// 	tsKey := r.generateInstance()

// 	slowlogs, err := redis.Values(c.Do("SLOWLOG", "GET", maxSlowEntries))
// 	if err != nil {
// 		return err
// 	}

// 	var maxTs int64
// 	for _, slowlog := range slowlogs {
// 		if entry, ok := slowlog.([]interface{}); ok {
// 			if entry == nil || len(entry) != 4 {
// 				return errors.New("slowlog get protocol error")
// 			}

// 			// id := entry[0].(int64)
// 			startTime := entry[1].(int64)
// 			if startTime <= r.lastTimestampSeen[tsKey] {
// 				continue
// 			}
// 			if startTime > maxTs {
// 				maxTs = startTime
// 			}
// 			duration := entry[2].(int64)

// 			var command []string
// 			if obj, ok := entry[3].([]interface{}); ok {
// 				for _, arg := range obj {
// 					command = append(command, string(arg.([]uint8)))
// 				}
// 			}

// 			commandTags := append(tags, "command:"+command[0])
// 			fields["slowlog_micros"] = duration
// 		}
// 	}
// 	r.lastTimestampSeen[tsKey] = maxTs
// 	return nil
// }

// func extractVal(s string) (val float64, err error) {
// 	split := strings.Split(s, "=")
// 	if len(split) != 2 {
// 		return 0, fmt.Errorf("nope")
// 	}
// 	val, err = strconv.ParseFloat(split[1], 64)
// 	if err != nil {
// 		return 0, fmt.Errorf("nope")
// 	}
// 	return
// }

// func extractConfig(config []string) (map[string]interface{}, error) {
// 	fields := make(map[string]interface{})

// 	if len(config)%2 != 0 {
// 		return nil, fmt.Errorf("invalid config: %#v", config)
// 	}

// 	for pos := 0; pos < len(config)/2; pos++ {
// 		val, err := strconv.ParseFloat(config[pos*2+1], 64)
// 		if err != nil {
// 			log.Debugf("couldn't parse %s, err: %s", config[pos*2+1], err)
// 			continue
// 		}
// 		fields[config[pos*2]] = val
// 	}
// 	return fields, nil
// }

func (r *Redis) Test() (*inputs.TestResult, error) {
	b.test = true
	b.resData = nil

	b.command()

	res := &inputs.TestResult{
		Result: b.resData,
		Desc:   "success!",
	}

	return res, nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Redis{}
	})
}

