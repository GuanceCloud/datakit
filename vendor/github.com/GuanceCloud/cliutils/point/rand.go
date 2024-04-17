// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:gosec
package point

import (
	"fmt"
	mrand "math/rand"
	"time"

	"github.com/GuanceCloud/cliutils"
)

type RandOption func(*ptRander)

type ptRander struct {
	pb,
	fixedKeys,
	fixedTags,
	kvSorted,
	stringValues bool

	ntext,
	ntags,
	nfields,
	keyLen,
	valLen int

	cat Category

	tagKeys    []string
	fieldKeys  []string
	tagVals    []string
	sampleText []string

	pointPool PointPool

	ts time.Time

	measurePrefix string
}

var defTags, defFields, defKeyLen, defValLen = 4, 10, 10, 17

func NewRander(opts ...RandOption) *ptRander {
	r := &ptRander{
		ntags:        defTags,
		nfields:      defFields,
		keyLen:       defKeyLen,
		valLen:       defValLen,
		ts:           time.Now(),
		cat:          Logging,
		stringValues: true,
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

	if r.ntext > 0 {
		if len(r.sampleText) == 0 {
			// nolint: lll
			r.sampleText = []string{
				`2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 10.000006916s ago...`,
				`2022-10-27T16:12:46.306+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a4xxxxxxxxxxxxxxxxxxx&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)`,
				`2022-10-27T16:12:46.306+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19dxxxxxxxxxxxxxxxxxxxxxxxx&filters=true)`,
				`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 2790747027482021869 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP and sampling ratio: 100%`,
				`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 1965248471827589152 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP and sampling ratio: 100%`,
				`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 2790747027482021869 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP.`,
				`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 1965248471827589152 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP.`,
				`2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/xxxxxx.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/1A552256-4134-4CAA-A4FF-7D2DEF11A6AC`,
				`2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/oss-browser.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/97346A30-EA8C-4AC8-991D-3AD64E2479E1`,
				`2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/Sublime Text.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/0EE2FB5D-6535-47AB-938B-DCB79CE11CE6`,
			}
		}
	}

	return r
}

func WithRandSampleText(st []string) RandOption     { return func(r *ptRander) { r.sampleText = st } }
func WithRandStringValues(on bool) RandOption       { return func(r *ptRander) { r.stringValues = on } }
func WithRandTime(t time.Time) RandOption           { return func(r *ptRander) { r.ts = t } }
func WithRandText(n int) RandOption                 { return func(r *ptRander) { r.ntext = n } }
func WithRandTags(n int) RandOption                 { return func(r *ptRander) { r.ntags = n } }
func WithRandFields(n int) RandOption               { return func(r *ptRander) { r.nfields = n } }
func WithRandKeyLen(n int) RandOption               { return func(r *ptRander) { r.keyLen = n } }
func WithRandValLen(n int) RandOption               { return func(r *ptRander) { r.valLen = n } }
func WithCategory(c Category) RandOption            { return func(r *ptRander) { r.cat = c } }
func WithRandPointPool(pp PointPool) RandOption     { return func(r *ptRander) { r.pointPool = pp } }
func WithRandMeasurementPrefix(s string) RandOption { return func(r *ptRander) { r.measurePrefix = s } }
func WithKVSorted(on bool) RandOption               { return func(r *ptRander) { r.kvSorted = on } }
func WithFixedKeys(on bool) RandOption              { return func(r *ptRander) { r.fixedKeys = on } }
func WithFixedTags(on bool) RandOption              { return func(r *ptRander) { r.fixedTags = on } }
func WithRandPB(on bool) RandOption                 { return func(r *ptRander) { r.pb = on } }

func (r *ptRander) Rand(count int) []*Point {
	if count <= 0 {
		return nil
	}

	pts := make([]*Point, 0, count)

	if pp := r.pointPool; pp != nil {
		SetPointPool(pp)
	}
	defer ClearPointPool()

	for i := 0; i < count; i++ {
		pts = append(pts, r.doRand())
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

func (r *ptRander) randTags() KVs {
	var kvs KVs
	for i := 0; i < r.ntags; i++ {
		var key, val string

		switch {
		case r.fixedTags:
			key = r.tagKeys[i]
			val = r.tagVals[i]

		case r.fixedKeys:
			key = r.tagKeys[i]
			val = randStr(r.valLen)
		default:
			key = randStr(r.keyLen)
			val = randStr(r.valLen)
		}

		if r.pointPool == nil {
			kvs = kvs.MustAddTag(key, val)
		} else {
			kv := r.pointPool.GetKV(key, val)
			kv.IsTag = true
			kvs = kvs.AddKV(kv, true)
		}
	}

	// add extra tags
	// nolint: exhaustive
	switch r.cat {
	case Object, CustomObject:
		// add `name` tag
		key, val := "name", randStr(r.valLen)

		if r.pointPool == nil {
			kvs = kvs.MustAddTag(key, val)
		} else {
			kv := r.pointPool.GetKV(key, val)
			kv.IsTag = true
			kvs = kvs.AddKV(kv, true)
		}
	default:
		// TODO:
	}

	return kvs
}

func (r *ptRander) randFields() KVs {
	var kvs KVs

	for i := 0; i < r.nfields; i++ {
		var (
			key = r.getFieldKey(i)
			val any
		)

		switch i {
		case 1:
			val = mrand.Float64()

		case 2:
			val = mrand.Float32()

		case 3, 4, 5:
			if r.stringValues {
				val = randStr(r.valLen)
			} else {
				continue
			}

		case 6, 7:
			val = (i%2 == 0) // bool

		default:
			val = mrand.Int63()
		}

		if r.pointPool == nil {
			kvs = kvs.Add(key, val, false, true) // force set field
		} else {
			kv := r.pointPool.GetKV(key, val)
			if kv == nil {
				panic(fmt.Sprintf("get nil kv on %q: %v", key, val))
			}
			kvs = kvs.AddKV(kv, true)
		}
	}

	// add long text field
	if len(r.sampleText) > 0 {
		for i := 0; i < r.ntext; i++ {
			key := "long-text" + randStr((i%r.keyLen)+1)
			val := r.sampleText[mrand.Int63()%int64(len(r.sampleText))]

			if r.pointPool == nil {
				kvs = kvs.Add(key, val, false, true)
			} else {
				kvs = kvs.AddKV(r.pointPool.GetKV(key, val), true)
			}
		}
	}

	return kvs
}

func (r *ptRander) doRand() *Point {
	var ptName string
	if r.measurePrefix != "" {
		ptName = r.measurePrefix + randStr(r.keyLen)
	} else {
		ptName = randStr(r.keyLen)
	}

	kvs := append(r.randTags(), r.randFields()...)
	pt := NewPointV2(ptName, kvs, WithTime(r.ts), WithKeySorted(r.kvSorted))

	if r.pb {
		pt.SetFlag(Ppb)
	}

	return pt
}
