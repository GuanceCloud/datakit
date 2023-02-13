// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:gosec
package point

import (
	mrand "math/rand"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
)

type RandOption func(*ptRander)

type ptRander struct {
	pb, fixedKeys, fixedTags              bool
	ntext, ntags, nfields, keyLen, valLen int

	cat Category

	tagKeys   []string
	fieldKeys []string
	tagVals   []string

	ts time.Time

	measurementPrefix string
}

var defTags, defFields, defKeyLen, defValLen = 4, 10, 10, 17

func NewRander(opts ...RandOption) *ptRander {
	r := &ptRander{
		ntags:   defTags,
		nfields: defFields,
		keyLen:  defKeyLen,
		valLen:  defValLen,
		ts:      time.Now(),
		cat:     Logging,
	}

	for _, opt := range opts {
		opt(r)
	}

	// add tag/field keys
	if r.fixedKeys {
		for i := 0; i < r.ntags; i++ {
			r.tagKeys = append(r.tagKeys, randStr(r.keyLen))
		}

		for i := 0; i < r.nfields; i++ {
			r.fieldKeys = append(r.fieldKeys, randStr(r.keyLen))
		}
	}

	if r.fixedTags {
		// add tag keys
		if len(r.tagKeys) == 0 {
			for i := 0; i < r.ntags; i++ {
				r.tagKeys = append(r.tagKeys, randStr(r.keyLen))
			}
		}

		// add tag values
		for range r.tagKeys {
			r.tagVals = append(r.tagVals, randStr(r.valLen))
		}
	}

	return r
}

func WithRandMeasurementPrefix(s string) RandOption {
	return func(r *ptRander) { r.measurementPrefix = s }
}
func WithRandTime(t time.Time) RandOption { return func(r *ptRander) { r.ts = t } }
func WithRandText(n int) RandOption       { return func(r *ptRander) { r.ntext = n } }
func WithRandTags(n int) RandOption       { return func(r *ptRander) { r.ntags = n } }
func WithRandFields(n int) RandOption     { return func(r *ptRander) { r.nfields = n } }
func WithRandKeyLen(n int) RandOption     { return func(r *ptRander) { r.keyLen = n } }
func WithRandValLen(n int) RandOption     { return func(r *ptRander) { r.valLen = n } }
func WithCategory(c Category) RandOption  { return func(r *ptRander) { r.cat = c } }

func WithFixedKeys(on bool) RandOption {
	return func(r *ptRander) {
		if on {
			r.fixedKeys = true
		} else {
			r.fixedKeys = false
		}
	}
}

func WithFixedTags(on bool) RandOption {
	return func(r *ptRander) {
		if on {
			r.fixedTags = true
		} else {
			r.fixedTags = false
		}
	}
}

func WithRandPB(on bool) RandOption {
	return func(r *ptRander) {
		if on {
			r.pb = true
		} else {
			r.pb = false
		}
	}
}

func (r *ptRander) Rand(count int) []*Point {
	if count <= 0 {
		return nil
	}

	var pts []*Point
	for i := 0; i < count; i++ {
		pt := r.doRand()
		if pt != nil {
			pts = append(pts, r.doRand())
		}
	}

	return pts
}

func randStr(n int) string {
	return cliutils.CreateRandomString(n)
}

func (r *ptRander) getFieldKey(i int) string {
	if r.fixedKeys {
		return r.fieldKeys[i]
	} else {
		return randStr(r.keyLen)
	}
}

func (r *ptRander) doRand() *Point {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	switch {
	case r.fixedTags:
		for i := 0; i < r.ntags; i++ { // reused exist tags
			tags[r.tagKeys[i]] = r.tagVals[i]
		}
	case r.fixedKeys:
		for i := 0; i < r.ntags; i++ { // reuse tag key but random tag value
			tags[r.tagKeys[i]] = randStr(r.valLen)
		}
	default: // random all tags (key & value)
		for i := 0; i < r.ntags; i++ {
			tags[randStr(r.keyLen)] = randStr(r.valLen)
		}
	}

	// add extra tags
	// nolint: exhaustive
	switch r.cat {
	case Object, CustomObject:
		// add `name` tag
		tags["name"] = randStr(r.valLen)

	default:
		// TODO:
	}

	for i := 0; i < r.nfields; i++ {
		switch i {
		case 1:
			fields[r.getFieldKey(i)] = mrand.Float64()
		case 2:
			fields[r.getFieldKey(i)] = mrand.Float32()
		case 3, 4, 5:
			fields[r.getFieldKey(i)] = randStr(r.valLen)
		case 6, 7:
			fields[r.getFieldKey(i)] = (i%2 == 0)

		default:
			fields[r.getFieldKey(i)] = mrand.Int63()
		}
	}

	n := r.ntext // currently only 3 reserved text field
	for {
		if n == 0 {
			break
		}

		fields["message"] = sampleLogs[mrand.Int63()%int64(len(sampleLogs))]
		n--

		if n == 0 {
			break
		}
		fields["error_message"] = sampleLogs[mrand.Int63()%int64(len(sampleLogs))]
		n--

		if n == 0 {
			break
		}
		fields["error_stack"] = sampleLogs[mrand.Int63()%int64(len(sampleLogs))]
		n--
	}

	measurement := r.measurementPrefix + randStr(r.keyLen)
	pt, err := NewPoint(measurement, tags, fields, WithTime(r.ts))
	if err != nil {
		return nil
	}

	if r.pb {
		pt.SetFlag(Ppb)
	}

	return pt
}

//nolint:lll
var (
	rawLogs = `
2022-10-27T16:12:54.876+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 971624677789410817 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP and sampling ratio: 100%
2022-10-27T16:12:54.876+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 564726768482716036 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP and sampling ratio: 100%
2022-10-27T16:12:54.876+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 971624677789410817 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP.
2022-10-27T16:12:54.876+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 564726768482716036 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP.
2022-10-27T16:12:54.876+0800	DEBUG	ddtrace	trace/aftergather.go:121	### send 2 points cost 0ms with error: <nil>
2022-10-27T16:12:54.875+0800	DEBUG	ddtrace	ddtrace/ddtrace_http.go:34	### received tracing data from path: /v0.4/traces
2022-10-27T16:12:54.281+0800	DEBUG	filter	filter/filter.go:158	filter condition body: {"dataways":null,"filters":{"logging":["{ source =  'datakit'  and ( host in [ 'ubt-dev-01' ,  'tanb-ubt-dev-test' ] )}"]},"pull_interval":10000000000,"remote_pipelines":null}
2022-10-27T16:12:54.184+0800	DEBUG	io	io/io.go:97	get iodata(1 points) from /v1/write/metric|swap
2022-10-27T16:12:54.184+0800	DEBUG	filter	filter/filter.go:408	update metrics...
2022-10-27T16:12:54.184+0800	DEBUG	filter	filter/filter.go:401	try pull remote filters...
2022-10-27T16:12:54.184+0800	DEBUG	filter	filter/filter.go:262	/v1/write/metric/pts: 1, after: 1
2022-10-27T16:12:54.184+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:54.184+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:54.183+0800	DEBUG	io	io/feed.go:91	io feed swap|/v1/write/metric
2022-10-27T16:12:54.183+0800	DEBUG	filter	filter/filter.go:235	no condition filter for metric
2022-10-27T16:12:53.688+0800	DEBUG	filter	filter/filter.go:158	filter condition body: {"dataways":null,"filters":{"logging":["{ source =  'datakit'  and ( host in [ 'ubt-dev-01' ,  'tanb-ubt-dev-test' ] )}"]},"pull_interval":10000000000,"remote_pipelines":null}
2022-10-27T16:12:53.622+0800	DEBUG	io	io/io.go:97	get iodata(2 points) from /v1/write/tracing|ddtrace
2022-10-27T16:12:53.622+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:49.573+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush dynamicDatawayCategory(0 pts), last flush 9.999510666s ago...
2022-10-27T16:12:49.462+0800	DEBUG	filter	filter/filter.go:158	filter condition body: {"dataways":null,"filters":{"logging":["{ source =  'datakit'  and ( host in [ 'ubt-dev-01' ,  'tanb-ubt-dev-test' ] )}"]},"pull_interval":10000000000,"remote_pipelines":null}
2022-10-27T16:12:49.389+0800	DEBUG	filter	filter/filter.go:401	try pull remote filters...
2022-10-27T16:12:49.389+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:49.389+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:49.388+0800	DEBUG	filter	filter/filter.go:408	update metrics...
2022-10-27T16:12:49.388+0800	DEBUG	filter	filter/filter.go:158	filter condition body: {"dataways":null,"filters":{"logging":["{ source =  'datakit'  and ( host in [ 'ubt-dev-01' ,  'tanb-ubt-dev-test' ] )}"]},"pull_interval":10000000000,"remote_pipelines":null}
2022-10-27T16:12:49.386+0800	DEBUG	io	io/io.go:97	get iodata(4 points) from /v1/write/tracing|ddtrace
2022-10-27T16:12:48.636+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:48.636+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:48.444+0800	DEBUG	filter	filter/filter.go:158	filter condition body: {"dataways":null,"filters":{"logging":["{ source =  'datakit'  and ( host in [ 'ubt-dev-01' ,  'tanb-ubt-dev-test' ] )}"]},"pull_interval":10000000000,"remote_pipelines":null}
2022-10-27T16:12:48.400+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:48.400+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:46.815+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:46.815+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:46.726+0800	DEBUG	filter	filter/filter.go:158	filter condition body: {"dataways":null,"filters":{"logging":["{ source =  'datakit'  and ( host in [ 'ubt-dev-01' ,  'tanb-ubt-dev-test' ] )}"]},"pull_interval":10000000000,"remote_pipelines":null}
2022-10-27T16:12:46.703+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=POST, string=url, *url.URL=https://openway.guance.com/v1/write/tracing?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588)
2022-10-27T16:12:46.700+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/write/tracing?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:46.699+0800	DEBUG	sender	sender/sender.go:47	sending /v1/write/object(1 pts)...
2022-10-27T16:12:46.699+0800	DEBUG	sender	sender/sender.go:47	sending /v1/write/metric(1 pts)...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:270	wal try flush failed data on /v1/write/security
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:270	wal try flush failed data on /v1/write/rum
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:270	wal try flush failed data on /v1/write/network
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:270	wal try flush failed data on /v1/write/metric
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:270	wal try flush failed data on /v1/write/logging
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/security(0 pts), last flush 10.000030625s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/rum(0 pts), last flush 9.999880583s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/rum(0 pts), last flush 9.999536583s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/rum(0 pts), last flush 9.999386542s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/rum(0 pts), last flush 9.999338708s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/rum(0 pts), last flush 9.998867333s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/rum(0 pts), last flush 9.998208209s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/object(1 pts), last flush 9.997395583s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/network(0 pts), last flush 9.99991425s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/network(0 pts), last flush 9.999568875s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/network(0 pts), last flush 9.998325375s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/network(0 pts), last flush 9.998172s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/network(0 pts), last flush 9.997431792s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(1 pts), last flush 9.999472083s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.999964541s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.999953542s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.999944333s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.999897792s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.999869417s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.999858791s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/metric(0 pts), last flush 9.99767025s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 9.999887125s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 9.998371916s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 9.997611625s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 9.997412708s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 10.002298833s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 10.000082958s ago...
2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 10.000006916s ago...
2022-10-27T16:12:46.306+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:46.306+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 2790747027482021869 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP and sampling ratio: 100%
2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 1965248471827589152 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP and sampling ratio: 100%
2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 2790747027482021869 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP.
2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 1965248471827589152 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP.
2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/网易有道词典.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/1A552256-4134-4CAA-A4FF-7D2DEF11A6AC
2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/oss-browser.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/97346A30-EA8C-4AC8-991D-3AD64E2479E1
2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/Sublime Text.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/0EE2FB5D-6535-47AB-938B-DCB79CE11CE6
2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/Microsoft Remote Desktop.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/DD10B11F-2D45-4DFD-B1CB-EF0F2B1FB2F7
2022-10-27T16:12:42.051+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 5484031498000114328 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP and sampling ratio: 100%
2022-10-27T16:12:42.051+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 1409415361793528756 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP and sampling ratio: 100%
2022-10-27T16:12:42.051+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 1409415361793528756 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP.
2022-10-27T16:12:42.051+0800	DEBUG	ddtrace	trace/aftergather.go:121	### send 2 points cost 0ms with error: <nil>
2022-10-27T16:12:42.051+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)
2022-10-27T16:12:42.051+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588&filters=true)
2022-10-27T16:12:42.050+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 5484031498000114328 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP.
2022-10-27T16:12:42.050+0800	DEBUG	ddtrace	ddtrace/ddtrace_http.go:34	### received tracing data from path: /v0.4/traces
	`
	sampleLogs = strings.Split(rawLogs, "\n")
)
