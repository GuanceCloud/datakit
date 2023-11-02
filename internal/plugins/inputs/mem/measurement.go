// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mem

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

type docMeasurement struct{}

// https://man7.org/linux/man-pages/man5/proc.5.html
// nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"total":             NewFieldInfoB("Total amount of memory."),
			"available":         NewFieldInfoB("Amount of available memory."),
			"available_percent": NewFieldInfoP("Available memory percent."),
			"used":              NewFieldInfoB("Amount of used memory."),
			"used_percent":      NewFieldInfoP("Used memory percent."),
			"active":            NewFieldInfoB("Memory that has been used more recently and usually not reclaimed unless absolutely necessary. (Darwin, Linux)"),
			"free":              NewFieldInfoB("Amount of free memory. (Darwin, Linux)"),
			"inactive":          NewFieldInfoB("Memory which has been less recently used.  It is more eligible to be reclaimed for other purposes. (Darwin, Linux)"),
			"wired":             NewFieldInfoB("Wired. (Darwin)"),
			"buffered":          NewFieldInfoB("Buffered. (Linux)"),
			"cached":            NewFieldInfoB("In-memory cache for files read from the disk. (Linux)"),
			"commit_limit":      NewFieldInfoB("This is the total amount of memory currently available to be allocated on the system. (Linux)"),
			"committed_as":      NewFieldInfoB("The amount of memory presently allocated on the system. (Linux)"),
			"dirty":             NewFieldInfoB("Memory which is waiting to get written back to the disk. (Linux)"),
			"high_free":         NewFieldInfoB("Amount of free high memory. (Linux)"),
			"high_total":        NewFieldInfoB("Total amount of high memory. (Linux)"),
			"huge_pages_free":   NewFieldInfoC("The number of huge pages in the pool that are not yet allocated. (Linux)"),
			"huge_pages_size":   NewFieldInfoB("The size of huge pages. (Linux)"),
			"huge_page_total":   NewFieldInfoC("The size of the pool of huge pages. (Linux)"),
			"low_free":          NewFieldInfoB("Amount of free low memory. (Linux)"),
			"low_total":         NewFieldInfoB("Total amount of low memory. (Linux)"),
			"mapped":            NewFieldInfoB("Files which have been mapped into memory, such as libraries. (Linux)"),
			"page_tables":       NewFieldInfoB("Amount of memory dedicated to the lowest level of page tables. (Linux)"),
			"shared":            NewFieldInfoB("Amount of shared memory. (Linux)"),
			"slab":              NewFieldInfoB("In-kernel data structures cache. (Linux)"),
			"sreclaimable":      NewFieldInfoB("Part of Slab, that might be reclaimed, such as caches. (Linux)"),
			"sunreclaim":        NewFieldInfoB("Part of Slab, that cannot be reclaimed on memory pressure. (Linux)"),
			"swap_cached":       NewFieldInfoB("Memory that once was swapped out, is swapped back in but still also is in the swap file. (Linux)"),
			"swap_free":         NewFieldInfoB("Amount of swap space that is currently unused. (Linux)"),
			"swap_total":        NewFieldInfoB("Total amount of swap space available. (Linux)"),
			"vmalloc_chunk":     NewFieldInfoB("Largest contiguous block of vmalloc area which is free. (Linux)"),
			"vmalloc_total":     NewFieldInfoB("Total size of vmalloc memory area. (Linux)"),
			"vmalloc_used":      NewFieldInfoB("Amount of vmalloc area which is used. (Linux)"),
			"write_back":        NewFieldInfoB("Memory which is actively being written back to the disk. (Linux)"),
			"write_back_tmp":    NewFieldInfoB("Memory used by FUSE for temporary write back buffers. (Linux)"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
	}
}
