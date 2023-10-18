// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"crypto/tls"
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func requestMetrics(url string) (map[string]*dto.MetricFamily, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint:gosec
	resp, err := http.Get(url)                                                                      // nolint:gosec
	if err != nil {
		return nil, err
	}

	var psr expfmt.TextParser

	mfs, err := psr.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	return mfs, err

	// ds := dkhttp.DatakitStats{
	//	DisableMonofont: true,
	//}
	// if err = json.Unmarshal(body, &ds); err != nil {
	//	return nil, err
	//}
}

/*
func requestInputInfo(url string) (map[string]string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}

	var info map[string]interface{}

	if err = json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	inputs := make(map[string]string, 1)
	for name, val := range info {
		bb, err := json.MarshalIndent(val, "", "\t")
		if err != nil {
			fmt.Println(err.Error())
			inputs[name] = err.Error()
			continue
		}
		inputs[name] = string(bb)
	}

	return inputs, nil
} */
