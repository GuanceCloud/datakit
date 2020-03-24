package prometheus

// import (
// 	"bufio"
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"log"
// 	"math"
// 	"mime"
// 	"net/http"
// 	"strings"
// 	"testing"
// 	"time"

// 	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

// 	"github.com/influxdata/telegraf"
// 	"github.com/influxdata/telegraf/metric"

// 	"github.com/matttproud/golang_protobuf_extensions/pbutil"
// 	dto "github.com/prometheus/client_model/go"
// 	"github.com/prometheus/common/expfmt"
// )

// const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,*/*;q=0.1`

// // Get Quantiles from summary metric
// func makeQuantiles(m *dto.Metric) map[string]interface{} {
// 	fields := make(map[string]interface{})
// 	for _, q := range m.GetSummary().Quantile {
// 		if !math.IsNaN(q.GetValue()) {
// 			fields[fmt.Sprint(q.GetQuantile())] = float64(q.GetValue())
// 		}
// 	}
// 	return fields
// }

// // Get Buckets  from histogram metric
// func makeBuckets(m *dto.Metric) map[string]interface{} {
// 	fields := make(map[string]interface{})
// 	for _, b := range m.GetHistogram().Bucket {
// 		fields[fmt.Sprint(b.GetUpperBound())] = float64(b.GetCumulativeCount())
// 	}
// 	return fields
// }

// func makeLabels(m *dto.Metric) map[string]string {
// 	result := map[string]string{}
// 	for _, lp := range m.Label {
// 		result[lp.GetName()] = lp.GetValue()
// 	}
// 	return result
// }

// func convertName(mf *dto.MetricFamily) (metricname string, valuename string) {
// 	rawname := mf.GetName()

// 	parts := strings.Split(rawname, "_")

// 	if len(parts) > 2 {
// 		metricname = strings.Join(parts[:2], "_")
// 		valuename = strings.Join(parts[2:], "_")
// 	} else {
// 		metricname = rawname
// 		valuename = "value"
// 	}
// 	return
// }

// // Get name and value from metric
// func getNameAndValue(m *dto.Metric, fieldname string) map[string]interface{} {
// 	fields := make(map[string]interface{})

// 	if m.Gauge != nil {
// 		if !math.IsNaN(m.GetGauge().GetValue()) {
// 			fields[fieldname] = float64(m.GetGauge().GetValue())
// 		}
// 	} else if m.Counter != nil {
// 		if !math.IsNaN(m.GetCounter().GetValue()) {
// 			fields[fieldname] = float64(m.GetCounter().GetValue())
// 		}
// 	} else if m.Untyped != nil {
// 		if !math.IsNaN(m.GetUntyped().GetValue()) {
// 			fields[fieldname] = float64(m.GetUntyped().GetValue())
// 		}
// 	}
// 	return fields
// }

// func valueType(mt dto.MetricType) telegraf.ValueType {
// 	switch mt {
// 	case dto.MetricType_COUNTER:
// 		return telegraf.Counter
// 	case dto.MetricType_GAUGE:
// 		return telegraf.Gauge
// 	case dto.MetricType_SUMMARY:
// 		return telegraf.Summary
// 	case dto.MetricType_HISTOGRAM:
// 		return telegraf.Histogram
// 	default:
// 		return telegraf.Untyped
// 	}
// }

// func parse(buf []byte, header http.Header) ([]telegraf.Metric, error) {

// 	var metrics []telegraf.Metric
// 	var parser expfmt.TextParser
// 	// parse even if the buffer begins with a newline
// 	buf = bytes.TrimPrefix(buf, []byte("\n"))
// 	// Read raw data
// 	buffer := bytes.NewBuffer(buf)
// 	reader := bufio.NewReader(buffer)

// 	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
// 	metricFamilies := make(map[string]*dto.MetricFamily)

// 	log.Printf("mediatype=%s, params=%v", mediatype, params)

// 	if err == nil && mediatype == "application/vnd.google.protobuf" &&
// 		params["encoding"] == "delimited" &&
// 		params["proto"] == "io.prometheus.client.MetricFamily" {
// 		for {
// 			mf := &dto.MetricFamily{}
// 			if _, ierr := pbutil.ReadDelimited(reader, mf); ierr != nil {
// 				if ierr == io.EOF {
// 					break
// 				}
// 				return nil, fmt.Errorf("reading metric family protocol buffer failed: %s", ierr)
// 			}
// 			log.Printf("%s", mf.GetName())
// 			metricFamilies[mf.GetName()] = mf
// 		}
// 	} else {
// 		metricFamilies, err = parser.TextToMetricFamilies(reader)
// 		if err != nil {
// 			return nil, fmt.Errorf("reading text format failed: %s", err)
// 		}
// 	}

// 	// read metrics
// 	for metricName, mf := range metricFamilies {
// 		_ = metricName
// 		for _, m := range mf.Metric {
// 			finalMetricName, valuename := convertName(mf)
// 			// reading tags
// 			tags := makeLabels(m)
// 			// reading fields
// 			fields := make(map[string]interface{})
// 			if mf.GetType() == dto.MetricType_SUMMARY {
// 				// summary metric
// 				fields = makeQuantiles(m)
// 				fields["count"] = float64(m.GetSummary().GetSampleCount())
// 				fields["sum"] = float64(m.GetSummary().GetSampleSum())
// 			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
// 				// histogram metric
// 				fields = makeBuckets(m)
// 				fields["count"] = float64(m.GetHistogram().GetSampleCount())
// 				fields["sum"] = float64(m.GetHistogram().GetSampleSum())

// 			} else {
// 				// standard metric
// 				fields = getNameAndValue(m, valuename)
// 			}
// 			// converting to telegraf metric
// 			if len(fields) > 0 {
// 				var t time.Time
// 				if m.TimestampMs != nil && *m.TimestampMs > 0 {
// 					t = time.Unix(0, *m.TimestampMs*1000000)
// 				} else {
// 					t = time.Now()
// 				}
// 				metric, err := metric.New(finalMetricName, tags, fields, t, valueType(mf.GetType()))
// 				if err == nil {
// 					log.Printf("%s", internal.Metric2InfluxLine(metric))
// 					metrics = append(metrics, metric)
// 				}
// 			}
// 		}
// 	}

// 	return metrics, err
// }

// func TestGetMetrics(t *testing.T) {

// 	cli := &http.Client{
// 		Transport: &http.Transport{
// 			DisableKeepAlives: true,
// 		},
// 	}

// 	req, err := http.NewRequest(http.MethodGet, "http://localhost:9100/metrics", nil)
// 	if err != nil {
// 		t.Errorf("%s", err)
// 	}

// 	req.Header.Add("Accept", acceptHeader)

// 	resp, err := cli.Do(req)
// 	if err != nil {
// 		t.Errorf("%s", err)
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("code=%v, %s", resp.StatusCode, resp.Status)
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Errorf("%s", err)
// 	}
// 	//log.Printf("%s", string(data))

// 	parse(body, resp.Header)
// }
