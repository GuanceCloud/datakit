// Copyright 2019-2023 VictoriaMetrics, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package prompb is used to unmarshal write requests
package prompb

// Reset resets wr.
func (wr *WriteRequest) Reset() { //nolint:stylecheck
	for i := range wr.Timeseries {
		ts := &wr.Timeseries[i]
		ts.Labels = nil
		ts.Samples = nil
	}
	wr.Timeseries = wr.Timeseries[:0]

	for i := range wr.labelsPool {
		lb := &wr.labelsPool[i]
		lb.Name = nil
		lb.Value = nil
	}
	wr.labelsPool = wr.labelsPool[:0]

	for i := range wr.samplesPool {
		s := &wr.samplesPool[i]
		s.Value = 0
		s.Timestamp = 0
	}
	wr.samplesPool = wr.samplesPool[:0]
}
