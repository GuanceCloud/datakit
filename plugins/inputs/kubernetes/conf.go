package kubernetes

const (
	configSample = `
[[inputs.kubernetes]]
    # required
    interval = "10s"

    ## URL for the Kubernetes API
    # url = "https://$HOSTIP:6443"

    ## Use bearer token for authorization.
    # bearer_token = "/path/to/token"
    ## CA file
    # tls_ca = "/path/to/ca_crt.pem"

    ## Set http timeout (default 5 seconds)
    # timeout = "5s"

    [inputs.kubernetes.tags]
    # tag1 = val1
    # tag2 = val2
`
)
