// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

func (x *candidate) pausePlugins() {
	defer func() {
		inputsPauseVec.WithLabelValues(x.id, x.namespace).Add(float64(len(x.plugins)))
	}()

	for i, p := range x.plugins {
		log.Debugf("pause %dth inputs...", i)
		if err := p.Pause(); err != nil {
			log.Warn(err)
		}
	}
}

func (x *candidate) resumePlugins() {
	defer func() {
		inputsResumeVec.WithLabelValues(x.id, x.namespace).Add(float64(len(x.plugins)))
	}()

	for i, p := range x.plugins {
		log.Debugf("resume %dth inputs...", i)
		if err := p.Resume(); err != nil {
			log.Warn(err)
		}
	}
}
