// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

const sampleConfig = `
[[inputs.ingestion_canary]]
  ## Collect interval, default 10m
  interval = "10m"

  ## Query timeout, should be less than interval, default 5m
  query_timeout = "5m"

  ## Poll interval for DQL query, default 500ms
  poll_interval = "500ms"

  ## Max retry count for query errors, default 10
  error_retries = 10

  ## Result workspace URL to report metrics
  # result_workspace = "https://openway.<<<custom_key.brand_main_domain>>>?token=xxx"

  ## Data categories to collect (metric, logging, tracing)
  ## If not specified, all categories will be collected
  categories = ["metric", "logging", "tracing"]

  ## Logging configuration
  [inputs.ingestion_canary.logging]
    storage_index = "default"

  ## Enable election mode
  election = true

  ## Extra tags
  [inputs.ingestion_canary.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
