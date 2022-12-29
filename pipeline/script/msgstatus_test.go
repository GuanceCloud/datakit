// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestStatus(t *testing.T) {
	for k, v := range statusMap {
		outp := &ptinput.Point{
			Fields: map[string]interface{}{
				FieldStatus: k,
			},
		}
		_, f, _ := ProcLoggingStatus(nil, outp.Fields, false, nil)
		assert.Equal(t, v, f[FieldStatus])
	}

	{
		outp := &ptinput.Point{
			Fields: map[string]interface{}{
				FieldStatus:  "x",
				FieldMessage: "1234567891011",
			},
		}
		_, f, _ := ProcLoggingStatus(nil, outp.Fields, false, nil)
		assert.Equal(t, "unknown", f[FieldStatus])
		assert.Equal(t, "1234567891011", f[FieldMessage])
	}

	{
		outp := &ptinput.Point{
			Fields: map[string]interface{}{
				FieldStatus:  "x",
				FieldMessage: "1234567891011",
			},
			Tags: map[string]string{
				"xxxqqqddd": "1234567891011",
			},
		}
		tags, f, _ := ProcLoggingStatus(outp.Tags, outp.Fields, false, nil)
		assert.Equal(t, map[string]interface{}{
			FieldStatus:  "unknown",
			FieldMessage: "1234567891011",
		}, f)
		assert.Equal(t, map[string]string{
			"xxxqqqddd": "1234567891011",
		}, tags)
	}

	{
		outp := &ptinput.Point{
			Fields: map[string]interface{}{
				FieldStatus:  "n",
				FieldMessage: "1234567891011",
			},
			Tags: map[string]string{
				"xxxqqqddd": "1234567891011",
			},
		}
		tags, f, _ := ProcLoggingStatus(outp.Tags, outp.Fields, false, nil)
		assert.Equal(t, map[string]interface{}{
			FieldStatus:  "notice",
			FieldMessage: "1234567891011",
		}, f)
		assert.Equal(t, map[string]string{
			"xxxqqqddd": "1234567891011",
		}, tags)
	}
}

func TestGetSetStatus(t *testing.T) {
	out := &ptinput.Point{
		Tags: map[string]string{
			"status": "n",
		},
		Fields: make(map[string]interface{}),
	}

	tags, f, _ := ProcLoggingStatus(out.Tags, out.Fields, false, nil)
	assert.Equal(t, map[string]string{
		"status": "notice",
	}, tags)
	assert.Equal(t, make(map[string]interface{}), f)

	out.Fields = map[string]interface{}{
		"status": "n",
	}
	out.Tags = make(map[string]string)
	tags, f, _ = ProcLoggingStatus(out.Tags, out.Fields, false, nil)
	assert.Equal(t, map[string]interface{}{
		"status": "notice",
	}, f)
	assert.Equal(t, make(map[string]string), tags)

	out.Fields = map[string]interface{}{
		"status": "n",
	}
	out.Tags = map[string]string{
		"status": "n",
	}
	tags, f, _ = ProcLoggingStatus(out.Tags, out.Fields, false, nil)
	assert.Equal(t, map[string]string{
		"status": "notice",
	}, tags)
	assert.Equal(t, map[string]interface{}{
		"status": "n",
	}, f)
}
