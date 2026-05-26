// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

// common prometheus summary objectives.
var (
	P8sStandardObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.01,
	}

	P8sLooseObjectives = map[float64]float64{
		0.5: 0.05,
	}

	p8sStrictObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}

	p8sHardObjectives = map[float64]float64{
		0.5:   0.05,
		0.9:   0.01,
		0.99:  0.001,
		0.999: 0.0001,
	}
)
