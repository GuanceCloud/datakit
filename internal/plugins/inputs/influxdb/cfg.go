// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package influxdb

const sampleConfig = `
[[inputs.influxdb]]
  url = "http://localhost:8086/debug/vars"

  ## (optional) collect interval, default is 10 seconds
  interval = '10s'
  
  ## Username and password to send using HTTP Basic Authentication.
  # username = ""
  # password = ""

  ## http request & header timeout
  timeout = "5s"

  ## Set true to enable election
  election = true

  ## TLS config
  # [inputs.influxdb.tlsconf]
    # insecure_skip_verify = true
    ## Following ca_certs/cert/cert_key are optional, if insecure_skip_verify = true.
    # ca_certs = ["/opt/tls/ca.crt"]
    # cert = "/opt/tls/client.root.crt"
    # cert_key = "/opt/tls/client.root.key"
    ## we can encode these file content in base64 format:
    # ca_certs_base64 = ["LONG_BASE64_STRING......"]
    # cert_base64 = "LONG_BASE64_STRING......"
    # cert_key_base64 = "LONG_BASE64_STRING......"
    # server_name = "your-SNI-name"

  # [inputs.influxdb.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "influxdb.p"

  [inputs.influxdb.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
