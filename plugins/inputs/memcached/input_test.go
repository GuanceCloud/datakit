package memcached

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var address = "localhost:11212"

func TestInput(t *testing.T) {
	memcached := &Input{}
	sampleMeasurements := memcached.SampleMeasurement()
	assert.Greater(t, len(sampleMeasurements), 0)
	m, ok := sampleMeasurements[0].(*inputMeasurement)
	if !ok {
		t.Error("expect *inputMeasurement")
	}

	assert.Equal(t, m.Info().Name, inputName)

	assert.Equal(t, memcached.Catalog(), catalogName)
	assert.Equal(t, memcached.SampleConfig(), sampleConfig)
	assert.Equal(t, memcached.AvailableArchs(), datakit.AllArch)
}

func TestParseMetrics(t *testing.T) {
	resp := bufio.NewReader(strings.NewReader(memcachedStats))
	values, err := parseResponse(resp)
	assert.Nil(t, err)

	checkValues(t, values)
}

func checkValues(t *testing.T, values map[string]string) {
	t.Helper()
	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("can't find key for metric %s in values", test.key)
		}
		if value != test.value {
			t.Errorf("metric: %s, Expected: %s, actual: %s", test.key, test.value, value)
		}
	}
}

func TestGatherServer(t *testing.T) {
	serverChan := make(chan int8)
	go func() {
		createTcpServer(t, serverChan)
	}()

	<-serverChan
	memcached := &Input{
		Servers: []string{address},
	}

	// err := memcached.gatherServer(address, false)
	err := memcached.Collect()
	assert.Nil(t, err)

	if len(memcached.collectCache) == 0 {
		assert.Fail(t, "collectCache is empty")
	}

	metric := memcached.collectCache[0]
	point, _ := metric.LineProto()
	assert.Equal(t, point.Name(), "memcached")
	fields, _ := point.Fields()
	values := make(map[string]string)
	for k, v := range fields {
		values[k] = fmt.Sprint(v)
	}
	checkValues(t, values)

	assert.NotNil(t, memcached.gatherServer(address, true))

	err = memcached.gatherServer("invalid url", false)
	assert.NotNil(t, err)
}

func createTcpServer(t *testing.T, serverChan chan<- int8) {
	t.Helper()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Errorf("mock tcp error: %s", err.Error())
	}
	serverChan <- 1
	for {
		conn, err := listener.Accept()
		if err != nil {
			assert.Fail(t, err.Error())
		}
		_, err = io.WriteString(conn, memcachedStats)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}
}

var memcachedStats = `STAT uptime 194
STAT curr_connections 5
STAT total_connections 6
STAT connection_structures 6
STAT cmd_get 0
STAT cmd_set 0
STAT cmd_flush 0
STAT cmd_touch 0
STAT get_hits 0
STAT get_misses 0
STAT delete_misses 0
STAT delete_hits 0
STAT incr_misses 0
STAT incr_hits 0
STAT decr_misses 0
STAT decr_hits 0
STAT cas_misses 0
STAT cas_hits 0
STAT cas_badval 0
STAT touch_hits 0
STAT touch_misses 0
STAT auth_cmds 0
STAT auth_errors 0
STAT bytes_read 7
STAT bytes_written 0
STAT limit_maxbytes 67108864
STAT accepting_conns 1
STAT listen_disabled_num 0
STAT conn_yields 0
STAT hash_power_level 16
STAT hash_bytes 524288
STAT hash_is_expanding 0
STAT expired_unfetched 0
STAT evicted_unfetched 0
STAT bytes 0
STAT curr_items 0
STAT total_items 0
STAT evictions 0
STAT reclaimed 0
END
`

var tests = []struct {
	key   string
	value string
}{
	{"uptime", "194"},
	{"curr_connections", "5"},
	{"total_connections", "6"},
	{"connection_structures", "6"},
	{"cmd_get", "0"},
	{"cmd_set", "0"},
	{"cmd_flush", "0"},
	{"cmd_touch", "0"},
	{"get_hits", "0"},
	{"get_misses", "0"},
	{"delete_misses", "0"},
	{"delete_hits", "0"},
	{"incr_misses", "0"},
	{"incr_hits", "0"},
	{"decr_misses", "0"},
	{"decr_hits", "0"},
	{"cas_misses", "0"},
	{"cas_hits", "0"},
	{"cas_badval", "0"},
	{"touch_hits", "0"},
	{"touch_misses", "0"},
	{"auth_cmds", "0"},
	{"auth_errors", "0"},
	{"bytes_read", "7"},
	{"bytes_written", "0"},
	{"limit_maxbytes", "67108864"},
	{"accepting_conns", "1"},
	{"listen_disabled_num", "0"},
	{"conn_yields", "0"},
	{"hash_power_level", "16"},
	{"hash_bytes", "524288"},
	{"hash_is_expanding", "0"},
	{"expired_unfetched", "0"},
	{"evicted_unfetched", "0"},
	{"bytes", "0"},
	{"curr_items", "0"},
	{"total_items", "0"},
	{"evictions", "0"},
	{"reclaimed", "0"},
}
