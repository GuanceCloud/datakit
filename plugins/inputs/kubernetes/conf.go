package kubernetes

const (
	configSample = `
[[inputs.kubernetes]]
    # required
    interval = "10s"

    ## URL for the Kubernetes API (required)
    # url = "https://$HOSTIP:6443"

    ## Use bearer token for authorization.(absolute path) (required)
    # bearer_token = "/path/to/token"

    ## CA file (absolute path) (required)
    # tls_ca = "/path/to/ca_crt.pem"

    ## Set http timeout (default 3 seconds)
    timeout = "3s"

    [inputs.kubernetes.tags]
    # tag1 = val1
    # tag2 = val2
`
)
