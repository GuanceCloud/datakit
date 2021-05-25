package docker_containers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultServiceAccountPath = "/run/secrets/kubernetes.io/serviceaccount/token"

// Kubernetes represents the config object for the plugin
type Kubernetes struct {
	URL string
	// Bearer Token authorization file path
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`
	ClientConfig

	roundTripper http.RoundTripper
}

func (k *Kubernetes) Init() error {
	// If neither are provided, use the default service account.
	if k.BearerToken == "" && k.BearerTokenString == "" {
		k.BearerToken = defaultServiceAccountPath
	}

	if k.BearerToken != "" {
		token, err := ioutil.ReadFile(k.BearerToken)
		if err != nil {
			return err
		}
		k.BearerTokenString = strings.TrimSpace(string(token))
	}

	return nil
}

func buildURL(endpoint string, base string) (*url.URL, error) {
	u := fmt.Sprintf(endpoint, base)
	addr, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse address '%s': %s", u, err)
	}
	return addr, nil
}

func (k *Kubernetes) GatherPodInfo(containerID string) (map[string]string, error) {
	var podApi Pods
	err := k.LoadJson(fmt.Sprintf("%s/pods", k.URL), &podApi)
	if err != nil {
		return nil, err
	}

	containerID = fmt.Sprintf("docker://%s", containerID)

	var m = make(map[string]string)

	for _, podMetadata := range podApi.Items {
		if len(podMetadata.Status.ContainerStatuses) == 0 {
			continue
		}
		for _, containerStauts := range podMetadata.Status.ContainerStatuses {
			if containerStauts.ContainerID == containerID {
				m["pod_name"] = podMetadata.Metadata.Name
				m["pod_namespace"] = podMetadata.Metadata.Namespace
				break
			}
		}
	}

	return m, nil
}

func (k *Kubernetes) LoadJson(url string, v interface{}) error {
	var req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	var resp *http.Response
	tlsCfg, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	if k.roundTripper == nil {
		k.roundTripper = &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			TLSClientConfig:       tlsCfg,
			ResponseHeaderTimeout: 5 * time.Second,
		}
	}
	req.Header.Set("Authorization", "Bearer "+k.BearerTokenString)
	req.Header.Add("Accept", "application/json")

	resp, err = k.roundTripper.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf(`Error parsing response: %s`, err)
	}

	return nil
}

type Pods struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	Items      []Item `json:"items"`
}

type Item struct {
	Metadata Metadata `json:"metadata"`
	Status   Status   `json:"status"`
}

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type Status struct {
	ContainerStatuses []ContainerStatus `json:"containerStatuses"`
}

type ContainerStatus struct {
	ContainerID string `json:"containerID"`
}
