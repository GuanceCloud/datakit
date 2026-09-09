// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"os"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
)

type upstreamGeoIPDB struct{}

func (*upstreamGeoIPDB) Init(string, map[string]string) {}
func (*upstreamGeoIPDB) SearchIsp(string) string        { return "unknown" }
func (db *upstreamGeoIPDB) GeoWithChecker(value string, check ipdb.CheckData) (*ipdb.IPdbRecord, error) {
	record, err := db.Geo(value)
	if record != nil && check != nil {
		record = check(record)
	}
	return record, err
}
func (db *upstreamGeoIPDB) Geo(value string) (*ipdb.IPdbRecord, error) {
	record := &ipdb.IPdbRecord{City: "Shanghai", Region: "Shanghai", Country: "CN", Isp: "unknown"}
	switch value {
	case "unknown-city":
		record.City = "unknown"
	case "unknown-region":
		record.Region = "unknown"
	case "unknown-country-short":
		record.Country = "unknown"
	}
	return record, nil
}

func TestGeoIPPipelineGoUpstreamCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 16,
		ServiceConfig{Version: 1}, func() (*HostCompat, error) {
			return NewPipelineGoHost(&upstreamGeoIPDB{}), nil
		})
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name, input, source string
		expected            map[string]string
		absent              []string
	}{
		{"plain", `{"ip":"1.2.3.4-something","second":2,"third":"abc","forth":true}`,
			`json(_, ip); geoip(ip)`, map[string]string{"city": "Shanghai", "country": "CN", "province": "Shanghai", "isp": "unknown"}, nil},
		{"unknown-city", `{"ip":"unknown-city"}`, `json(_, ip); geoip(ip)`,
			map[string]string{"city": "unknown", "country": "CN", "province": "Shanghai", "isp": "unknown"}, nil},
		{"attribute", `{"aa":{"ip":"116.228.89.xxx"}}`, `json(_, aa.ip); geoip(aa.ip)`,
			map[string]string{"city": "Shanghai", "country": "CN", "province": "Shanghai", "isp": "unknown"}, nil},
		{"unknown-region", `{"aa":{"ip":"unknown-region"}}`, `json(_, aa.ip); geoip(aa.ip)`,
			map[string]string{"city": "Shanghai", "country": "CN", "province": "unknown", "isp": "unknown"}, nil},
		{"unknown-country", `{"aa":{"ip":"unknown-country-short"}}`, `json(_, aa.ip); geoip(aa.ip)`,
			map[string]string{"city": "Shanghai", "country": "unknown", "province": "Shanghai", "isp": "unknown"}, nil},
		{"literal-prefix", `{"ip":"1.2.3.4-something"}`, `json(_, ip); add_key(city,"existing"); geoip(ip,"geo_")`,
			map[string]string{"city": "existing", "geo_city": "Shanghai", "geo_country": "CN", "geo_province": "Shanghai", "geo_isp": "unknown"}, []string{"country", "province", "isp"}},
		{"dynamic-prefix", `{"ip":"1.2.3.4-something"}`, `json(_, ip); p="geo_"; geoip(ip,p)`,
			map[string]string{"geo_city": "Shanghai", "geo_country": "CN", "geo_province": "Shanghai", "geo_isp": "unknown"}, nil},
		{"named-prefix", `{"ip":"1.2.3.4-something"}`, `json(_, ip); geoip(ip,prefix="geo_")`,
			map[string]string{"geo_city": "Shanghai", "geo_country": "CN", "geo_province": "Shanghai", "geo_isp": "unknown"}, nil},
		{"empty-prefix", `{"ip":"1.2.3.4-something"}`, `json(_, ip); geoip(ip,"")`,
			map[string]string{"city": "Shanghai", "country": "CN", "province": "Shanghai", "isp": "unknown"}, nil},
		{"missing-ip", `{"second":2}`, `geoip(ip,"geo_"); add_key(after,true)`, map[string]string{}, []string{"geo_city", "geo_country", "geo_province", "geo_isp"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := runner.Check(tc.source)
			if check.Route != RouteJITWithHost || check.Capabilities.Backend != ExecutionBackendMachineCode ||
				check.Capabilities.RequiredHostFlags == 0 || len(check.Capabilities.HostCalls) != 1 || check.Capabilities.HostCalls[0] != "geoip" {
				t.Fatalf("route=%+v", check)
			}
			for _, size := range []int{1, 2, 4, 8, 10, 128} {
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{"message": tc.input, "sequence": int64(i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("batch %d records=%d error=%v", size, len(batch.Records), err)
				}
				for i := range batch.Records {
					if batch.Records[i].Status != TerminalOK {
						t.Fatalf("batch %d record %d=%+v", size, i, batch.Records[i])
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					for key, want := range tc.expected {
						if points[i].Fields[key] != want {
							t.Fatalf("batch %d record %d key %s=%#v want=%q: %#v", size, i, key, points[i].Fields[key], want, points[i])
						}
					}
					for _, key := range tc.absent {
						if _, ok := points[i].Fields[key]; ok {
							t.Fatalf("batch %d record %d unexpected key %s: %#v", size, i, key, points[i])
						}
					}
					if points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("batch %d record %d lost input: %#v", size, i, points[i])
					}
				}
			}
		})
	}
}

func TestGeoIPPipelineGoInvalidPrefixCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 4,
		ServiceConfig{Version: 1}, func() (*HostCompat, error) { return NewPipelineGoHost(&upstreamGeoIPDB{}), nil })
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	for _, source := range []string{`geoip(ip,2)`, `geoip(ip,"a","b")`} {
		if check := runner.Check(source); check.Route != RoutePipelineGo || check.Reason != CheckReasonCompileError {
			t.Fatalf("source %q route=%+v", source, check)
		}
	}
	const dynamic = `add_key(before,true); p=123; geoip(ip,p); add_key(after,true)`
	if check := runner.Check(dynamic); check.Route != RouteJITWithHost {
		t.Fatalf("dynamic route=%+v", check)
	}
	point := Point{Version: 1, Category: "logging", Measurement: "geoip-error",
		Fields: map[string]any{"ip": "1.2.3.4-something"}}
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(dynamic, input)
	if err != nil || len(batch.Records) != 1 {
		t.Fatalf("batch=%+v error=%v", batch, err)
	}
	record := batch.Records[0]
	if record.Status != TerminalError || !record.CommitPrefixError || !record.HasMutations() || len(record.Error) == 0 {
		t.Fatalf("record=%+v", record)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if point.Fields["before"] != true || point.Fields["after"] != nil {
		t.Fatalf("committed/stopped output=%#v", point)
	}
}
