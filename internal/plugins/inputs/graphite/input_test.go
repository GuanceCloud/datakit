// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package graphite integrate test
package graphite

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestParse(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		cfg := `
		address = ":9109"
		
		[metric_mapper]
		  name = "test"
		  [[metric_mapper.mappings]]
		  match = "test.dispatcher.*.*.*"
		  name = "dispatcher_events_total"
		
			[metric_mapper.mappings.labels]
			action = "$2"
			job = "test_dispatcher"
			outcome = "$3"
			processor = "$1"
		
		  [[metric_mapper.mappings]]
		  match = "*.signup.*.*"
		  name = "signup_events_total"
		
			[metric_mapper.mappings.labels]
			job = "${1}_server"
			outcome = "$3"
			provider = "$2"
		
		  [[metric_mapper.mappings]]
		  match = '''servers\.(.*)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.*)'''
		  match_type = "regex"
		
			[metric_mapper.mappings.labels]
			hostname = "$1"
			device = "$2"
		`

		// Initialize the default input struct
		n := defaultInput()

		// Decode the TOML configuration into the struct
		if _, err := toml.Decode(cfg, n); err != nil {
			t.Log("Error decoding TOML:", err)
			return
		}

		t.Logf("Parsed config: %+v\n", n)

		// Check if mappings are populated
		if len(n.MetricMapper.Mappings) == 0 {
			t.Log("No mappings found.")
		} else {
			for i, mapping := range n.MetricMapper.Mappings {
				t.Logf("Mapping %d: %+v\n", i+1, mapping)
			}
		}
	})
}

/*
func TestLoad(t *testing.T) {
	cfg := `
	[[inputs.graphite]]
  ## Address to open UDP/TCP, default 9109
  address = ":9109"


  [inputs.graphite.metric_mapper]
  name = "test"
  [[inputs.graphite.metric_mapper.mappings]]
  match = "test.dispatcher2.*.*.*"
  name = "dispatcher2_events_total"

  [inputs.graphite.metric_mapper.mappings.labels]
  action = "$2"
  job = "test_dispatcher"
  outcome = "$3"
  processor = "$1"

  [[inputs.graphite.metric_mapper.mappings]]
  match = "*.signup.*.*"
  name = "signup_events_total"

  [inputs.graphite.metric_mapper.mappings.labels]
  job = "${1}_server"
  outcome = "$3"
  provider = "$2"

  [[inputs.graphite.metric_mapper.mappings]]
  match = '''servers\.(.*)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.*)'''
  match_type = "regex"
  name = 'servers_networking_transmissions_${3}'

  [inputs.graphite.metric_mapper.mappings.labels]
  hostname = "$1"
  device = "$2"
  # strict_match = false
	`
	ipts := map[string]inputs.Creator{}
	f := func() inputs.Input {
		return defaultInput()
	}
	ipts["graphite"] = f
	pts, err := config.LoadSingleConf(cfg, ipts)
	if err != nil {
		t.Fatal(err)
	}
	graphiteInput := pts["graphite"][0]
	graphiteInput.Run()
	fmt.Printf("ipts: %v", pts)
}
*/
