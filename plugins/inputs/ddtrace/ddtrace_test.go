package ddtrace

import (
	"encoding/base64"
	"io"
	"math/rand"
	"testing"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

type debugDDTraceMock struct {
	count         int
	spanGenerator func(count int) [][]*Span
}

func (this *debugDDTraceMock) unmarshalDdtraceMsgpack(body io.ReadCloser) ([][]*Span, error) {
	return this.spanGenerator(this.count), nil
}

func (this *debugDDTraceMock) statistic(origin [][]*Span, sampled []*dkio.Point) {
	log.Info("####### sample statistic")
	for k, conf := range sampleConfs {
		log.Infof("rules[%d]: %v", k, *conf)
	}

	scope := 0
	for _, spans := range origin {
		scope += len(spans)
	}
	origErr := 0
	for _, span := range origin[0] {
		if span.Error > 0 {
			origErr++
		}
	}
	sampledErr := 0
	for _, point := range sampled {
		if point.Tags()[trace.TAG_SPAN_STATUS] == trace.STATUS_ERR {
			sampledErr++
		}
	}
	rate := len(sampled)
	log.Infof("sample rate: %d%%", int(float64(rate)/float64(scope)*100))
	log.Infof("error in origin: %d", origErr)
	log.Infof("error in sampled: %d", sampledErr)
	log.Infof("sample rate without errors: %d%%", int(float64(rate-sampledErr)/float64(scope-origErr)*100))
}

func randomSpan() *Span {
	// meta := make(map[string]string, 10)
	// for i := 0; i < len(meta); i++ {
	// 	meta[randomName(6)] = randomName(9)
	// }
	// metrics := make(map[string]float64, 10)
	// for i := 0; i < len(metrics); i++ {
	// 	metrics[randomName(6)] = rand.Float64()
	// }

	return &Span{
		Name:     randomName(9),
		Service:  randomName(9),
		Resource: randomName(9),
		Type:     randomName(6),
		Start:    int64(randomId(11)),
		Duration: int64(randomId(11)),
		// Meta:     meta,
		// Metrics:  metrics,
		SpanID:   randomId(10),
		TraceID:  randomId(10),
		ParentID: randomId(10),
	}
}

func randomName(l int) string {
	if l <= 0 {
		l = 10
	}
	var buf = make([]byte, l)
	rand.Read(buf)

	return base64.StdEncoding.EncodeToString(buf)
}

func randomId(l int) uint64 {
	if l == 0 {
		return 0
	}

	var num float64 = float64(rand.Intn(9)+1) + rand.Float64()
	for i := 1; i < l; i++ {
		num *= 10
	}

	return uint64(num)
}

func TestDDTraceSampleWithNoErrorNoIgnore(t *testing.T) {
	sampleConfs = []*trace.TraceSampleConfig{
		&trace.TraceSampleConfig{
			Target: map[string]string{"name": "zhuyun"},
			Rate:   9,
			Scope:  100,
		},
		&trace.TraceSampleConfig{
			Target: map[string]string{"age": "123"},
			Rate:   18,
			Scope:  100,
		},
		&trace.TraceSampleConfig{
			Rate:  27,
			Scope: 100,
		},
	}
	defDDTraceMock = &debugDDTraceMock{
		count: 1000,
		spanGenerator: func(count int) [][]*Span {
			var spans []*Span
			for j := 0; j < count; j++ {
				span := randomSpan()
				span.Error = 0
				spans = append(spans, span)
			}

			return [][]*Span{spans}
		}}
	for i := 0; i < 30; i++ {
		if err := parseDdtraceMsgpack(nil); err != nil {
			panic(err.Error())
		}
	}
}

func TestDDTraceSampleWithError(t *testing.T) {
	sampleConfs = []*trace.TraceSampleConfig{
		&trace.TraceSampleConfig{
			Target: map[string]string{"name": "zhuyun"},
			Rate:   9,
			Scope:  100,
		},
		&trace.TraceSampleConfig{
			Target: map[string]string{"age": "123"},
			Rate:   18,
			Scope:  100,
		},
		&trace.TraceSampleConfig{
			Rate:  27,
			Scope: 100,
		},
	}
	defDDTraceMock = &debugDDTraceMock{
		count: 1000,
		spanGenerator: func(count int) [][]*Span {
			var spans []*Span
			for j := 0; j < count; j++ {
				span := randomSpan()
				if rand.Intn(10) >= 5 {
					span.Error = 1
				}
				spans = append(spans, span)
			}

			return [][]*Span{spans}
		}}
	for i := 0; i < 30; i++ {
		if err := parseDdtraceMsgpack(nil); err != nil {
			panic(err.Error())
		}
	}
}

func TestDDTraceSampleWithIgnoreTags(t *testing.T) {
	sampleConfs = []*trace.TraceSampleConfig{
		&trace.TraceSampleConfig{
			Target: map[string]string{"name": "zhuyun"},
			Rate:   9,
			Scope:  100,
		},
		&trace.TraceSampleConfig{
			Target: map[string]string{"age": "123"},
			Rate:   18,
			Scope:  100,
		},
		&trace.TraceSampleConfig{
			Rate:           27,
			Scope:          100,
			IgnoreTagsList: []string{trace.PROJECT},
		},
	}
	defDDTraceMock = &debugDDTraceMock{
		count: 1000,
		spanGenerator: func(count int) [][]*Span {
			var spans []*Span
			for j := 0; j < count; j++ {
				span := randomSpan()
				// if rand.Intn(10) >= 5 {
				// 	span.Meta = map[string]string{"tnt": "this is ignore tag"}
				// }
				spans = append(spans, span)
			}

			return [][]*Span{spans}
		}}
	for i := 0; i < 30; i++ {
		if err := parseDdtraceMsgpack(nil); err != nil {
			panic(err.Error())
		}
	}
}
