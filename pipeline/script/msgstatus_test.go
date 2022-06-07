package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func TestStatus(t *testing.T) {
	for k, v := range statusMap {
		outp := &parser.Output{
			Fields: map[string]interface{}{
				FieldStatus: k,
			},
		}
		outp = ProcLoggingStatus(outp, false, nil)
		assert.Equal(t, v, outp.Fields[FieldStatus])
	}

	{
		outp := &parser.Output{
			Fields: map[string]interface{}{
				FieldStatus:  "x",
				FieldMessage: "1234567891011",
			},
		}
		outp = ProcLoggingStatus(outp, false, nil)
		assert.Equal(t, "unknown", outp.Fields[FieldStatus])
		assert.Equal(t, "1234567891011", outp.Fields[FieldMessage])
	}

	{
		outp := &parser.Output{
			Fields: map[string]interface{}{
				FieldStatus:  "x",
				FieldMessage: "1234567891011",
			},
			Tags: map[string]string{
				"xxxqqqddd": "1234567891011",
			},
		}
		outp = ProcLoggingStatus(outp, false, nil)
		assert.Equal(t, map[string]interface{}{
			FieldStatus:  "unknown",
			FieldMessage: "1234567891011",
		}, outp.Fields)
		assert.Equal(t, map[string]string{
			"xxxqqqddd": "1234567891011",
		}, outp.Tags)
	}

	{
		outp := &parser.Output{
			Fields: map[string]interface{}{
				FieldStatus:  "n",
				FieldMessage: "1234567891011",
			},
			Tags: map[string]string{
				"xxxqqqddd": "1234567891011",
			},
		}
		outp = ProcLoggingStatus(outp, false, nil)
		assert.Equal(t, map[string]interface{}{
			FieldStatus:  "notice",
			FieldMessage: "1234567891011",
		}, outp.Fields)
		assert.Equal(t, map[string]string{
			"xxxqqqddd": "1234567891011",
		}, outp.Tags)
	}
}

func TestGetSetStatus(t *testing.T) {
	out := &parser.Output{
		Tags: map[string]string{
			"status": "n",
		},
		Fields: make(map[string]interface{}),
	}

	out = ProcLoggingStatus(out, false, nil)
	assert.Equal(t, map[string]string{
		"status": "notice",
	}, out.Tags)
	assert.Equal(t, make(map[string]interface{}), out.Fields)

	out.Fields = map[string]interface{}{
		"status": "n",
	}
	out.Tags = make(map[string]string)
	out = ProcLoggingStatus(out, false, nil)
	assert.Equal(t, map[string]interface{}{
		"status": "notice",
	}, out.Fields)
	assert.Equal(t, make(map[string]string), out.Tags)

	out.Fields = map[string]interface{}{
		"status": "n",
	}
	out.Tags = map[string]string{
		"status": "n",
	}
	out = ProcLoggingStatus(out, false, nil)
	assert.Equal(t, map[string]string{
		"status": "notice",
	}, out.Tags)
	assert.Equal(t, map[string]interface{}{
		"status": "n",
	}, out.Fields)
}
