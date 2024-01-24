// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"fmt"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/point"
)

const httpMetricName = "host_http_exception"

func (ipt *Input) collectHTTP() error {
	ts := time.Now()

	for _, http := range ipt.http {
		statusCode := fmt.Sprintf("%d", http.ExpectStatus)
		for _, url := range http.HTTPURLs {
			task := dt.HTTPTask{
				Method:     http.Method,
				URL:        url,
				ExternalID: "-",
				SuccessWhen: []*dt.HTTPSuccess{
					{
						StatusCode: []*dt.SuccessOption{
							{
								Is: statusCode,
							},
						},
					},
				},
				AdvanceOptions: &dt.HTTPAdvanceOption{
					Certificate: &dt.HTTPOptCertificate{
						IgnoreServerCertificateError: http.IgnoreInsecureTLS,
					},
					RequestOptions: &dt.HTTPOptRequest{
						Headers: http.Headers,
					},
					RequestTimeout: http.Timeout,
				},
			}
			if err := task.InitDebug(); err != nil {
				l.Warnf("init http task failed: %s", err.Error())
				continue
			}

			if err := task.Run(); err != nil {
				l.Warnf("run http task failed: %s", err.Error())
			}

			tags, fields := task.GetResults()

			if tags["status"] == "FAIL" {
				if reason, ok := fields["fail_reason"].(string); ok {
					var kvs point.KVs
					kvs = kvs.Add("url", url, true, true)
					kvs = kvs.Add("error", reason, true, true)
					kvs = kvs.Add("exception", true, false, true)

					for k, v := range ipt.mergedTags {
						kvs = kvs.AddTag(k, v)
					}

					opts := point.DefaultMetricOptions()
					opts = append(opts, point.WithTime(ts))

					ipt.collectCache = append(ipt.collectCache, point.NewPointV2(httpMetricName, kvs, opts...))
				}
			}
		}
	}

	return nil
}
