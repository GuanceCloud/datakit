// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (

	// nolint:gosec
	"crypto/md5"
	"database/sql"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ConversionFunc func(value sql.RawBytes) (interface{}, error)

func Conv(val interface{}, datatype string) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	switch datatype {
	case inputs.Float:
		res, err = cast.ToFloat64E(val)
	case inputs.Int:
		switch x := val.(type) {
		case int:
			return int64(x), nil
		case int8:
			return int64(x), nil
		case int16:
			return int64(x), nil
		case int32:
			return int64(x), nil
		case int64:
			return x, nil

		case uint:
			return uint64(x), nil
		case uint8:
			return uint64(x), nil
		case uint16:
			return uint64(x), nil
		case uint32:
			return uint64(x), nil
		case uint64:
			return x, nil

		case string:
			if i64, err := strconv.ParseInt(x, 10, 64); err == nil {
				return i64, nil
			} else if u64, err := strconv.ParseUint(x, 10, 64); err == nil {
				return u64, nil
			} else {
				return val, fmt.Errorf("unknown int value")
			}

		default:
			return val, fmt.Errorf("unknown int value, should not been here")
		}
	case inputs.Bool:
		res, err = cast.ToBoolE(val)
	case inputs.String:
		res, err = cast.ToStringE(val)
	}

	return res, err
}

type mysqlVersion struct {
	version string
	flavor  string
	build   string
}

func (m *mysqlVersion) versionCompatible(compatVersion []int) bool {
	const LEN = 3
	currentVersion := [LEN]int{0, 0, 0}
	targetVersion := []int{0, 0, 0}
	var err error

	copy(targetVersion, compatVersion)

	versions := strings.Split(m.version, ".")

	if len(versions) != 3 {
		return false
	}

	if currentVersion[0], err = strconv.Atoi(versions[0]); err != nil {
		return false
	}

	if currentVersion[1], err = strconv.Atoi(versions[1]); err != nil {
		return false
	}

	patchPattern := regexp.MustCompile(`(\d+).*`)
	res := patchPattern.FindStringSubmatch(versions[2])

	if len(res) > 1 {
		if currentVersion[2], err = strconv.Atoi(res[1]); err != nil {
			return false
		}
	} else {
		return false
	}

	for i := 0; i < LEN; i++ {
		if currentVersion[i] != targetVersion[i] {
			return currentVersion[i] > targetVersion[i]
		}
	}

	return true
}

func canExplain(obfuscatedStatement string) bool {
	parts := strings.Split(obfuscatedStatement, " ")
	if len(parts) < 2 {
		return false
	}
	op := parts[0]
	for _, keyword := range []string{"select", "table", "delete", "insert", "replace", "update", "with"} {
		if strings.ToLower(op) == keyword {
			return true
		}
	}
	return false
}

var reg = regexp.MustCompile(`\n|\s+`) //nolint:gocritic

func obfuscateSQL(text string) (sql string) {
	defer func() {
		sql = strings.TrimSpace(reg.ReplaceAllString(sql, " "))
	}()

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", text); err != nil {
		return fmt.Sprintf("ERROR: failed to obfuscate: %s", err.Error())
	} else {
		return out.Query
	}
}

func computeSQLSignature(text string) (signature string) {
	signature = fmt.Sprintf("%x", md5.Sum([]byte(text))) //nolint:gosec
	return
}

// TODO: obfuscate,.
func obfuscatePlan(text string) string {
	reg := regexp.MustCompile(`\n\s+`)
	out := strings.TrimSpace(reg.ReplaceAllString(text, ""))
	// remove \n
	return strings.ReplaceAll(out, "\n", "")
}

type cacheItem struct {
	expire time.Time
}

// simple ttl cache, only for database monitoring cache.
type cacheLimit struct {
	Size      int
	TTL       int64
	itemStore map[string]cacheItem
}

func (c *cacheLimit) len() int {
	count := 0
	for k := range c.itemStore {
		if ok := c.get(k); ok {
			count++
		}
	}

	return count
}

func (c *cacheLimit) get(key string) bool {
	if value, ok := c.itemStore[key]; ok {
		if time.Now().Before(value.expire) {
			return true
		} else {
			delete(c.itemStore, key)
			return false
		}
	}
	return false
}

func (c *cacheLimit) add(key string) {
	if c.itemStore == nil {
		c.itemStore = make(map[string]cacheItem)
	}
	duration := time.Duration(c.TTL) * time.Second
	expire := time.Now().Add(duration)
	c.itemStore[key] = cacheItem{
		expire: expire,
	}
}

func (c *cacheLimit) Acquire(key string) bool {
	if c.len() >= c.Size {
		return false
	}
	if ok := c.get(key); !ok {
		c.add(key)
		return true
	}
	return false
}

func isListHasStr(list []string, str string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

func getHostTag(host string) string {
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		return host
	}

	return ""
}

func getMetricNames(name string) (string, string) {
	names := strings.SplitN(name, ":", 2)
	metricName := ""
	sqlName := ""
	if len(names) == 1 {
		metricName = names[0]
		sqlName = names[0]
	} else if len(names) == 2 {
		metricName = names[0]
		sqlName = names[1]
	}

	return metricName, sqlName
}

func getMetricName(metricName, sqlName string) string {
	if sqlName == "" {
		return metricName
	} else {
		return metricName + ":" + sqlName
	}
}
