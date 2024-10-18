// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/GuanceCloud/cliutils"
)

var (
	flagLen   = flag.Int64("len", 32, "generated data length(kb)")
	flagCount = flag.Int("count", 3, "generated data count")
	flagP8s   = flag.Bool("p8s", false, "generate promethues metric text")
	flagFile  = flag.String("output", "v1.data", "data output to file")
)

func genLargeLog() {
	fd, err := os.OpenFile(*flagFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < *flagCount; i++ {
		n := *flagLen * 1024

		if i > 0 { // append \n to previous data
			if _, err := fd.WriteString("\n"); err != nil {
				panic(err.Error())
			}
		}

		for {
			if n/1024 > 0 { // each time generate 1kb data
				if _, err := fd.WriteString(cliutils.CreateRandomString(1024)); err != nil {
					panic(err.Error())
				}
				n -= 1024
			} else {
				if _, err := fd.WriteString(cliutils.CreateRandomString(int(n % 1024))); err != nil {
					panic(err.Error())
				}

				break
			}
		}
	}
}

func genLargeP8sMetric() {
	countComment := `# HELP some counter
# TYPE my_count counter`
	countMetricTemplate := `my_count{tag1="%05d",tag2="%05d"} 5.739356555004705
`

	gaugeComment := `# HELP some gauge
# TYPE my_gauge gauge`

	gaugeMetricTemplate := `my_gauge{tag1="%05d",tag2="%05d"} 5.739356555004705
`

	summaryComment := `# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary`

	summaryMetricTemplate := `go_gc_duration_seconds{tag1="%05d",tag2="%05d",quantile="0"} 0.000159
go_gc_duration_seconds{tag1="%05d",tag2="%05d",quantile="0.25"} 0.000346833
go_gc_duration_seconds{tag1="%05d",tag2="%05d",quantile="0.5"} 0.000542084
go_gc_duration_seconds{tag1="%05d",tag2="%05d",quantile="0.75"} 0.000859208
go_gc_duration_seconds{tag1="%05d",tag2="%05d",quantile="1"} 0.010519458
go_gc_duration_seconds_sum{tag1="%05d",tag2="%05d"} 0.76161121
go_gc_duration_seconds_count{tag1="%05d",tag2="%05d"} 1234
`

	histogramComment := `# HELP go_gc_heap_allocs_by_size_bytes Distribution of heap allocations by approximate size. Note that this does not include tiny objects as defined by /gc/heap/tiny/allocs:objects, only tiny blocks.
# TYPE go_gc_heap_allocs_by_size_bytes histogram`

	histogramMetricTemplate := `go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="8.999999999999998"} 1.9150407e+07
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="24.999999999999996"} 2.72283031e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="64.99999999999999"} 3.66184003e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="144.99999999999997"} 4.1888465e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="320.99999999999994"} 4.30278691e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="704.9999999999999"} 4.3276057e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="1536.9999999999998"} 4.33416083e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="3200.9999999999995"} 4.33837422e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="6528.999999999999"} 4.34562466e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="13568.999999999998"} 4.3479673e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="27264.999999999996"} 4.34931507e+08
go_gc_heap_allocs_by_size_bytes_bucket{tag1="%05d",tag2="%05d",le="+Inf"} 4.35160484e+08
go_gc_heap_allocs_by_size_bytes_sum{tag1="%05d",tag2="%05d"} 2.325903862e+11
go_gc_heap_allocs_by_size_bytes_count{tag1="%05d",tag2="%05d"} 4.35160484e+08
`

	// generate n count point
	fmt.Println(countComment)
	for i := 0; i < *flagCount; i++ {
		fmt.Printf(countMetricTemplate, i, i)
	}

	// generate n gauge point
	fmt.Println(gaugeComment)
	for i := 0; i < *flagCount; i++ {
		fmt.Printf(gaugeMetricTemplate, i, i)
	}

	// generate n summary point
	fmt.Println(summaryComment)
	for i := 0; i < *flagCount; i++ {
		fmt.Printf(summaryMetricTemplate, i, i, i, i, i, i, i, i, i, i, i, i, i, i)
	}

	// generate n histogram point
	fmt.Println(histogramComment)
	for i := 0; i < *flagCount; i++ {
		fmt.Printf(histogramMetricTemplate, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i)
	}
}

// nolint: typecheck
func main() {
	flag.Parse()
	if *flagP8s {
		genLargeP8sMetric()
		return
	}
	genLargeLog()
}
