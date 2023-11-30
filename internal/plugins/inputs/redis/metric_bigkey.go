// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	redisHotkey      = "redis_hotkey"
	redisBigkey      = "redis_bigkey"
	redisClient      = "redis_client"
	redisCluster     = "redis_cluster"
	redisCommandStat = "redis_command_stat"
	redisDB          = "redis_db"
	redisLatency     = "redis_latency"
	redisInfoM       = "redis_info"
	redisReplica     = "redis_replica"
	redisSlowlog     = "redis_slowlog"
)

type bigKeyMeasurement struct{}

//nolint:lll
func (m *bigKeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisBigkey,
		Type: "logging",
		Fields: map[string]interface{}{
			"value_length": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Key length."},
			"keys_sampled": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Sampled keys in the key space."},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname."},
			"server":       &inputs.TagInfo{Desc: "Server addr."},
			"service_name": &inputs.TagInfo{Desc: "Service name."},
			"db_name":      &inputs.TagInfo{Desc: "DB name."},
			"key":          &inputs.TagInfo{Desc: "Key name."},
			"key_type":     &inputs.TagInfo{Desc: "Key type."},
		},
	}
}

func (ipt *Input) collectBigKey() error {
	keys, err := ipt.getKeys()
	if err != nil {
		return err
	}

	pts, err := ipt.getData(keys)
	if err != nil {
		return err
	}

	if len(pts) > 0 {
		err = ipt.feeder.Feed(redisBigkey, point.Logging, pts, &io.Option{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) getKeys() ([]string, error) {
	var res []string
	for _, pattern := range ipt.Keys {
		var cursor uint64
		for {
			var keys []string
			var err error
			ctx := context.Background()

			keys, cursor, err = ipt.client.Scan(ctx, cursor, pattern, 10).Result()
			if err != nil {
				l.Errorf("redis pattern key %s scan fail error %v", pattern, err)
				return nil, err
			}
			// keys: []string{"key1","key2"...}

			res = append(res, keys...)
			if cursor == 0 {
				break
			}
		}
	}
	return res, nil
}

func (ipt *Input) getData(resKeys []string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))

	for _, key := range resKeys {
		var kvs point.KVs
		kvs = kvs.AddTag("db_name", fmt.Sprintf("%d", ipt.DB))
		kvs = kvs.AddTag("key", key)

		found := false
		ctx := context.Background()
		for _, op := range []string{
			"HLEN",
			"LLEN",
			"SCARD",
			"ZCARD",
			"PFCOUNT",
			"STRLEN",
		} {
			val, err := ipt.client.Do(ctx, op, key).Result()
			if err == nil && val != nil {
				// op:"STRLEN", key:"key1", val=interface{}(int64)5
				found = true
				kvs = kvs.Add("value_length", val, false, true)
				break
			}
		}

		if !found {
			if ipt.WarnOnMissingKeys {
				l.Warnf("%s key not found in redis", key)
			}

			kvs = kvs.Add("value_length", 0, false, true)
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisBigkey, kvs, opts...))
		}
	}

	return collectCache, nil
}

func (ipt *Input) goroutineBigKey(ctxKey context.Context) {
	g := datakit.G("redis_bigkey")
	g.Go(func(ctx context.Context) error {
		tickBigKey := time.NewTicker(ipt.KeyInterval)
		defer tickBigKey.Stop()

		for {
			if !ipt.pause {
				if err := ipt.scanBigKey(ctxKey); err != nil {
					l.Errorf("scanBigKey: %s", err)
				}
			}

			select {
			case <-datakit.Exit.Wait():
				ipt.exit()
				l.Info("redis exit")
				return nil

			case <-ipt.semStop.Wait():
				ipt.exit()
				l.Info("redis return")
				return nil

			case <-tickBigKey.C:
			case ipt.pause = <-ipt.pauseCh:
				// nil
			}
		}
	})
}

func (ipt *Input) scanBigKey(ctxKey context.Context) error {
	for _, db := range ipt.keyDBS {
		data, err := ipt.getBigData(ctxKey, db)
		if err != nil {
			return err
		}

		pts, err := ipt.parseBigData(data, db)
		if err != nil {
			return err
		}

		if len(pts) > 0 {
			err = ipt.feeder.Feed(redisBigkey, point.Logging, pts, &io.Option{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ipt *Input) getBigData(ctxKey context.Context, db int) (string, error) {
	// ctx create from ctxKey, cancelKey() when Run() func end.
	ctx, cancel := context.WithTimeout(ctxKey, ipt.KeyTimeout)
	defer cancel()

	args := []string{"redis-cli", "--bigkeys", "-i", ipt.KeyScanSleep}
	if ipt.Username != "" && ipt.Password != "" {
		// See also: https://redis.io/docs/connect/cli/
		// Example: redis-cli --hotkeys -i 0.1 -u redis://LJenkins:p%40ssw0rd@192.168.0.2:6379/0
		u := url.QueryEscape(ipt.Username) + ":" + url.QueryEscape(ipt.Password)
		u = "redis://" + u + "@" + ipt.Host + ":" + fmt.Sprint(ipt.Port) + "/" + fmt.Sprint(db)
		args = append(args, "-u", u)
	} else {
		args = append(args, "-h", ipt.Host, "-p", fmt.Sprint(ipt.Port))
	}
	//nolint:gosec
	c := exec.CommandContext(ctx, args[0], args[1:]...)

	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return "", fmt.Errorf("c.Start(): %w, %v", err, b.String())
	}
	if err := c.Wait(); err != nil {
		return "", fmt.Errorf("c.Wait(): %s, %w, %v", inputName, err, b.String())
	}

	bytes := b.Bytes()
	l.Debugf("get bytes len: %v.", len(bytes))

	return string(bytes), nil
}

func (ipt *Input) parseBigData(data string, db int) ([]*point.Point, error) {
	rdr := strings.NewReader(data)
	scanner := bufio.NewScanner(rdr)

	collectCache := []*point.Point{}
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))

	for scanner.Scan() {
		// Example: [52.76%] Biggest string found so far '"keySlice2999"' with 47984 bytes (not useful)
		// Example: Sampled 3006 keys in the keyspace!
		// Example: Biggest   hash found '"myhash"' has 2 fields
		// Example: Biggest string found '"keySlice2999"' has 47984 bytes
		// Example: Biggest   zset found '"keyZSet2999"' has 3001 members
		line := scanner.Text()
		var kv map[string]interface{}
		message := "big key "

		if strings.HasPrefix(line, "Sampled ") &&
			strings.HasSuffix(line, " keys in the keyspace!") {
			kv = getSampled(line)
			if v, ok := kv["keys_sampled"]; ok {
				message += " keys_sampled: " + fmt.Sprint(v)
			}
		}

		if strings.HasPrefix(line, "Biggest ") &&
			strings.Contains(line, " found '\"") &&
			strings.Contains(line, "\"' has ") {
			kv = getBigKey(line)
			if v, ok := kv["key"]; ok {
				message += " key: " + fmt.Sprint(v)
			}
			if v, ok := kv["key_type"]; ok {
				message += " key_type: " + fmt.Sprint(v)
			}
			if v, ok := kv["value_length"]; ok {
				message += " value_length: " + fmt.Sprint(v)
			}
		}

		if len(kv) == 0 {
			continue
		}

		var kvs point.KVs
		kvs = kvs.AddTag("db_name", strconv.Itoa(db))
		for k, v := range kv {
			kvs = kvs.Add(k, v, false, false)
		}
		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		kvs = kvs.Add("message", message, false, false)

		collectCache = append(collectCache, point.NewPointV2(redisBigkey, kvs, opts...))
	}

	return collectCache, nil
}

func getBigKey(line string) map[string]interface{} {
	// Example: Biggest   hash found '"myhash"' has 2 fields
	// Example: Biggest string found '"keySlice2999"' has 47984 bytes
	// Example: Biggest   zset found '"keyZSet2999"' has 3001 members
	kv := map[string]interface{}{}

	line = strings.TrimPrefix(line, "Biggest ")
	line = strings.ReplaceAll(line, "\"' has ", " found '\"")
	parts := strings.Split(line, " found '\"")
	if len(parts) != 3 {
		return kv
	}

	bigKeyCounterStrs := strings.Split(parts[2], " ")
	valueLength, err := strconv.Atoi(bigKeyCounterStrs[0])
	if err != nil {
		return kv
	}

	kv["value_length"] = valueLength
	kv["key"] = strings.Trim(parts[1], " ")
	kv["key_type"] = strings.Trim(parts[0], " ")

	return kv
}
