// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

const (
	inputName = "prom_remote_write"
	catalog   = "prom"
	sample    = `
[[inputs.prom_remote_write]]
  ## Path to listen to.
  path = "/prom_remote_write"

  ## accepted methods
  methods = ["PUT", "POST"]
  
  ## If the data is decoded incorrectly, you need to set the default HTTP body encoding;
  ## this usually occurs when the sender does not correctly pass the encoding in the HTTP header.
  #
  # default_content_encoding = "snappy"

  ## Part of the request to consume.  Available options are "body" and "query".
  # data_source = "body"

  ## output source
  # specify this to output collected metrics to local file
  # if not specified, metrics is sent to datakit io
  # if specified, you can use 'datakit --prom-conf /path/to/this/conf' to debug collected data
  # output = "/abs/path/file"

  ## If job_as_measurement is true, use the job field for the measurement name.
  ## The measurement_name configuration takes precedence.
  job_as_measurement = false

  ## Metric name filter
  # Regex is supported.
  # Only metric matches one of the regex can pass through. No filter if left empty.
  # metric_name_filter = ["gc", "go"]

  ## Measurement name filter
  # Regex is supported.
  # Only measurement matches one of the regex can pass through. No filter if left empty.
  # This filtering is done before any prefixing rule or renaming rule is applied.
  # measurement_name_filter = ["kubernetes", "container"]

  ## metric name prefix
  ## prefix will be added to metric name
  # measurement_prefix = "prefix_"

  ## metric name
  ## metric name will be divided by "_" by default.
  ## metric is named by the first divided field, the remaining field is used as the current metric name
  ## metric name will not be divided if measurement_name is configured
  ## measurement_prefix will be added to the start of measurement_name
  # measurement_name = "prom_remote_write"

  ## max body size in bytes, default set to 500MB
  # max_body_size = 0

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # basic_username = ""
  # basic_password = ""

  ## If both blacklist and whitelist, all list will cancel.
  ## tags to ignore (blacklist)
  # tags_ignore = ["xxxx"]

  ## tags to ignore with regex (blacklist)
  # tags_ignore_regex = ["xxxx"]

  ## tags whitelist
  # tags_only = ["xxxx"]

  ## tags whitelist with regex
  # tags_only_regex = ["xxxx"]

  ## Indicate whether tags_rename overwrites existing key if tag with the new key name already exists.
  overwrite = false

  ## tags to rename
  [inputs.prom_remote_write.tags_rename]
  # old_tag_name = "new_tag_name"
  # more_old_tag_name = "other_new_tag_name"

  ## Optional setting to map http headers into tags
  ## If the http header is not present on the request, no corresponding tag will be added
  ## If multiple instances of the http header are present, only the first value will be used
  [inputs.prom_remote_write.http_header_tags]
  # HTTP_HEADER = "TAG_NAME"

  ## Customize measurement set name.
  ## Treat those metrics with prefix as one set.
  ## Prioritier over 'measurement_name' configuration.
  ## Must measurement_name = ""
  [[inputs.prom_remote_write.measurements]]
    prefix = "etcd_network_"
    name = "etcd_network"
    
  [[inputs.prom_remote_write.measurements]]
    prefix = "etcd_server_"
    name = "etcd_server"

  ## custom tags
  [inputs.prom_remote_write.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

// defaultMaxBodySize is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 500 MB.
const defaultMaxBodySize int64 = 500 * 1024 * 1024
