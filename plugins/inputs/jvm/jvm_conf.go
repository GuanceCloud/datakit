package jvm

const (
	javaConfSample = `#[[inputs.jvm]]
  ## Add agents URLs to query
  # urls = "http://localhost:8080/jolokia"
  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Monitor Intreval
  # interval   = "60s"
`
)
