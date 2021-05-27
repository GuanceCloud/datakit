package kubernetes

const (
	configSample = `
[[inputs.kubernetes]]
    # required
    interval = "10s"

    ## URL for the Kubernetes API
    # url = "https://$HOSTIP:6443"

    ## Use bearer token for authorization. ('bearer_token' takes priority)
    ## at: /run/secrets/kubernetes.io/serviceaccount/token
    # bearer_token_string = "abc_123"

    ## Set http timeout (default 5 seconds)
    # timeout = "5s"

    ## Optional TLS Config
    # tls_ca = "/path/to/cafile"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false

    [inputs.kubernetes.apiservice]
    #  url = ""

    [inputs.kubernetes.tags]
    # tag1 = val1
    # tag2 = val2
`
)
