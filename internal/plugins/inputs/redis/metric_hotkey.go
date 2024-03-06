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
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type hotkeyMeasurement struct{}

//nolint:lll
func (m *hotkeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisHotkey,
		Type: "logging",
		Fields: map[string]interface{}{
			"key_count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Key count times."},
			"keys_sampled": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Sampled keys in the key space."},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname."},
			"server":       &inputs.TagInfo{Desc: "Server addr."},
			"service_name": &inputs.TagInfo{Desc: "Service name."},
			"db_name":      &inputs.TagInfo{Desc: "DB name."},
			"key":          &inputs.TagInfo{Desc: "Key name."},
		},
	}
}

func (ipt *Input) goroutineHotkey(ctxKey context.Context) {
	g := datakit.G("redis_hotkey")
	g.Go(func(ctx context.Context) error {
		tickHotkey := time.NewTicker(ipt.KeyInterval)
		defer tickHotkey.Stop()

		for {
			if !ipt.pause {
				if err := ipt.scanHotkey(ctxKey); err != nil {
					l.Errorf("scanHotkey: %s", err)
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

			case <-tickHotkey.C:
			case ipt.pause = <-ipt.pauseCh:
				// nil
			}
		}
	})
}

func (ipt *Input) scanHotkey(ctxKey context.Context) error {
	for _, db := range ipt.keyDBS {
		data, err := ipt.getHotData(ctxKey, db)
		if err != nil {
			return err
		}

		pts, err := ipt.parseHotData(data, db)
		if err != nil {
			return err
		}

		if len(pts) > 0 {
			if err := ipt.feeder.FeedV2(point.Logging, pts,
				dkio.WithElection(ipt.Election),
				dkio.WithInputName(redisHotkey)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ipt *Input) getHotData(ctxKey context.Context, db int) (string, error) {
	// ctx create from ctxKey, cancelKey() when Run() func end.
	ctx, cancel := context.WithTimeout(ctxKey, ipt.KeyTimeout)
	defer cancel()

	args := []string{"redis-cli", "--hotkeys", "-i", ipt.KeyScanSleep}
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

func (ipt *Input) parseHotData(data string, db int) ([]*point.Point, error) {
	rdr := strings.NewReader(data)
	scanner := bufio.NewScanner(rdr)

	collectCache := []*point.Point{}
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))

	for scanner.Scan() {
		// Example: [39.22%] Hot key 'keySlice70' found so far with counter 5 (not useful)
		// Example: Sampled 100002 keys in the keyspace!
		// Example: hot key found with counter: 5\tkeyname: "keySlice5316"
		line := scanner.Text()
		var kv map[string]interface{}
		message := "hot key "

		if strings.HasPrefix(line, "Sampled ") &&
			strings.HasSuffix(line, " keys in the keyspace!") {
			kv = getSampled(line)
			if v, ok := kv["keys_sampled"]; ok {
				message += " keys_sampled: " + fmt.Sprint(v)
			}
		}

		if strings.HasPrefix(line, "hot key found with counter:") {
			kv = getHotkey(line)
			if v, ok := kv["key"]; ok {
				message += " key: " + fmt.Sprint(v)
			}
			if v, ok := kv["key_count"]; ok {
				message += " key_count: " + fmt.Sprint(v)
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

		collectCache = append(collectCache, point.NewPointV2(redisHotkey, kvs, opts...))
	}

	return collectCache, nil
}

func getSampled(line string) map[string]interface{} {
	// Example: Sampled 100002 keys in the keyspace!
	kv := map[string]interface{}{}

	line = strings.TrimPrefix(line, "Sampled ")
	line = strings.TrimSuffix(line, " keys in the keyspace!")

	hotkeySampled, err := strconv.Atoi(line)
	if err != nil {
		return kv
	}

	kv["keys_sampled"] = hotkeySampled

	return kv
}

func getHotkey(line string) map[string]interface{} {
	// Example: hot key found with counter: 5\tkeyname: "keySlice5316"
	kv := map[string]interface{}{}

	line = strings.TrimPrefix(line, "hot key found with counter:")
	line = strings.ReplaceAll(line, "\t", " ")
	parts := strings.Split(line, "keyname:")
	if len(parts) != 2 {
		return kv
	}

	hotkeyCounter, err := strconv.Atoi(strings.Trim(parts[0], " "))
	if err != nil {
		return kv
	}

	kv["key_count"] = hotkeyCounter
	kv["key"] = strings.Trim(strings.Trim(parts[1], " "), "\"")

	return kv
}
