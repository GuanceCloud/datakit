// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"fmt"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/point"
)

const httpMetricName = "host_http_exception"

func (ipt *Input) collectHTTP(ptTS int64) error {
	for _, http := range ipt.http {
		statusCode := fmt.Sprintf("%d", http.ExpectStatus)
		for _, url := range http.HTTPURLs {
			ct := &dt.HTTPTask{
				Task: &dt.Task{
					ExternalID: "-",
				},
				Method: http.Method,
				URL:    url,
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
			task, err := dt.NewTask("", ct)
			if err != nil {
				l.Warnf("newTask failed: %s", err.Error())
				continue
			}

			if err := task.RenderTemplateAndInit(nil); err != nil {
				l.Warnf("init http task failed: %s", err.Error())
				continue
			}

			if err := task.Run(); err != nil {
				l.Warnf("run http task failed: %s", err.Error())
			}

			tags, fields := task.GetResults()
			var kvs point.KVs
			kvs = kvs.SetTag("url", url)
			kvs = kvs.Set("exception", false)
			kvs = kvs.SetTag("error", "none")

			if tags["status"] == "FAIL" {
				kvs = kvs.Set("exception", true)
				if reason, ok := fields["fail_reason"].(string); ok {
					kvs = kvs.SetTag("error", reason)
				}
			}

			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}

			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTimestamp(ptTS))

			ipt.collectCache = append(ipt.collectCache, point.NewPoint(httpMetricName, kvs, opts...))
		}
	}

	return nil
}
