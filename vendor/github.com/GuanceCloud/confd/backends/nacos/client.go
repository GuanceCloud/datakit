package nacos

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	BackendNodes   []BackendNode
	Password       string
	Username       string
	Namespace      string
	AccessKey      string
	SecretKey      string
	CircleInterval int
}

type BackendNode struct {
	BackendAddr   string
	NamespaceName string
	AccessToken   string
}

// NewNacosClient create nacos client.
func NewNacosClient(
	backendNodes []string,
	password,
	username,
	namespace,
	accessKey,
	secretKey string,
	circleInterval int,
) (*Client, error) {
	c := &Client{}
	for _, backendNode := range backendNodes {
		if _, err := url.ParseRequestURI(backendNode); err != nil {
			return nil, fmt.Errorf("nacos backendNode url Parse : %v", err)
		}
		c.BackendNodes = append(c.BackendNodes, BackendNode{backendNode, "", ""}) // deep copy
	}
	c.Password = password
	c.Username = username
	c.Namespace = namespace
	c.AccessKey = accessKey
	c.SecretKey = secretKey
	c.CircleInterval = circleInterval
	if c.CircleInterval < 1 {
		c.CircleInterval = 60
	}

	// Check password.
	if c.Password != "" {
		if err := c.loginWithPassword(); err != nil {
			return nil, fmt.Errorf("nacos login with password error : %v", err)
		}
	}

	// Get namespace name.
	for i := 0; i < len(c.BackendNodes); i++ {
		namespaceName, err := getNamespaceName(c.BackendNodes[i].BackendAddr, c.Namespace)
		if err != nil {
			return nil, err
		}
		c.BackendNodes[i].NamespaceName = namespaceName
	}
	return c, nil
}

type RspBody struct {
	AccessToken string `json:"accessToken"`
	TokenTtl    int    `json:"tokenTtl"`
	GlobalAdmin bool   `json:"globalAdmin"`
	Username    string `json:"username"`
}

func (c *Client) loginWithPassword() error {
	// Nacos POST body must this way.
	reqBodyStr := fmt.Sprintf("username=%s&password=%s", c.Username, c.Password)
	reqBody := strings.NewReader(reqBodyStr)
	for i := 0; i < len(c.BackendNodes); i++ {
		url := c.BackendNodes[i].BackendAddr + "/nacos/v1/auth/login"
		httpReq, err := http.NewRequest("POST", url, reqBody)
		if err != nil {
			// log.Error("NewRequest fail, url: %s, reqBody: %s, err: %v", url, reqBody, err)
			return err
		}
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		httpReq.Header.Add("Content-Type", "application/json")

		// Do HTTP.
		httpRsp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			// log.Error("do http fail, url: %s, reqBody: %s, err:%v", url, reqBody, err)
			return err
		}
		defer httpRsp.Body.Close()

		// Read HTTP response.
		rspBody, err := ioutil.ReadAll(httpRsp.Body)
		if httpRsp.StatusCode != 200 {
			// log.Error("response err: %v", string(rspBody))
			return fmt.Errorf("response err: %v", string(rspBody))
		}
		if err != nil {
			// log.Error("ReadAll failed, url: %s, reqBody: %s, err: %v", url, reqBody, err)
			return err
		}

		// Unmarshal response body.
		var result RspBody
		if err = json.Unmarshal(rspBody, &result); err != nil {
			// log.Error("Unmarshal fail, err:%v", err)
			return err
		}

		// log.Debug("do post http success, url: %s, reqBody: %s, body: %s %s", url, reqBody, string(rspBody), errMsg)
		c.BackendNodes[i].AccessToken = result.AccessToken
	}

	return nil
}

type NamespaceResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    []NamespacePageItem `json:"data"`
}
type NamespacePageItem struct {
	Namespace         string `json:"namespace"`
	NamespaceShowName string `json:"namespaceShowName"`
	NamespaceDesc     string `json:"namespaceDesc"`
	Quota             int    `json:"quota"`
	ConfigCount       int    `json:"configCount"`
	Type              int    `json:"type"`
}

func getNamespaceName(backendAddr, namespaceID string) (string, error) {
	url := backendAddr + `/nacos/v1/console/namespaces`
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	obj := NamespaceResponse{}
	err = json.Unmarshal(b, &obj)
	if err != nil {
		_ = err
	}

	// Find the right namespace name.
	for i := 0; i < len(obj.Data); i++ {
		if obj.Data[i].Namespace == namespaceID {
			return obj.Data[i].NamespaceShowName, nil
		}
	}

	return "", fmt.Errorf("not find this namespace id: %s", namespaceID)
}

// GetValues get all dataID prefix in this namespace.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("nacos GetValues keys is nil")
	}

	// To return map[string]string.
	values := make(map[string]string)

	// Search data.
	for _, backendNode := range c.BackendNodes {
		err := c.getValuesFromHTTP(backendNode.BackendAddr, backendNode.NamespaceName, backendNode.AccessToken, keys, values)
		if err != nil {
			if strings.HasPrefix(err.Error(), "code not 200") {
				// HTTP code not 200, retry login.
				// log.Error("HTTP code not 200, retry login.", err)
				if errLogin := c.loginWithPassword(); errLogin != nil {
					return make(map[string]string), fmt.Errorf("nacos login with password error in getValues: %v", errLogin)
				}
			}
			return make(map[string]string), fmt.Errorf("nacos configClient.SearchConfig : %v %v", backendNode, err)
		}
	}

	return values, nil
}

type ValuesResponse struct {
	TotalCount     int              `json:"totalCount"`
	PageNumber     int              `json:"pageNumber"`
	PagesAvailable int              `json:"pagesAvailable"`
	PageItems      []ValuesPageItem `json:"pageItems"`
}
type ValuesPageItem struct {
	ID               string `json:"id"`
	DataId           string `json:"dataId"`
	Group            string `json:"group"`
	Content          string `json:"content"`
	Md5              string `json:"md5"`
	EncryptedDataKey string `json:"encryptedDataKey"`
	Tenant           string `json:"tenant"`
	AppName          string `json:"appName"`
	Type             string `json:"type"`
}

func (c *Client) getValuesFromHTTP(backendAddr, namespaceName, accessToken string, keys []string, values map[string]string) error {
	url := backendAddr + `/nacos/v1/cs/configs?pageSize=65535&group=&dataId=&search=blur&pageNo=1&tenant=` + c.Namespace
	if accessToken != "" {
		url = url + "&accessToken=" + accessToken
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("code not 200 err : %v ", string(b))
	}
	if err != nil {
		return err
	}

	obj := ValuesResponse{}
	err = json.Unmarshal(b, &obj)
	if err != nil {
		_ = err
	}

	// Get only with prefix in keys.
	for i := 0; i < len(obj.PageItems); i++ {
		keyPath := namespaceName + "/" + obj.PageItems[i].Group + "/" + obj.PageItems[i].DataId
		for j := 0; j < len(keys); j++ {
			if strings.HasPrefix(keyPath, keys[j]) {
				values[keyPath] = obj.PageItems[i].Content
				break
			}
		}
	}

	return nil
}

// WatchPrefix watch all namespace.
// If creat new dataID, perhaps delay 60S.
// @prefix @keys all useful.
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	prefixes := append([]string{}, keys...)
	if prefix != "" {
		prefixes = append(prefixes, prefix)
	}
	// List of all dataIDs prefix, the key like "/group_name/dataID_name".
	oldDataIDs, err := c.GetValues(prefixes)
	if err != nil {
		return waitIndex + 1, fmt.Errorf("nacos WatchPrefix : %v", err)
	}

	tick := time.NewTicker(time.Second * time.Duration(c.CircleInterval))
	defer tick.Stop()
	for {
		select {
		case <-stopChan:
			return waitIndex, fmt.Errorf("stopChan done")
		case <-tick.C:
		}

		// Cycle get dataIDs prefix to find modify.
		newDataIDs, err := c.GetValues(append(keys, prefix))
		if err != nil {
			return waitIndex + 1, fmt.Errorf("nacos WatchPrefix : %v", err)
		}
		// Deep equal.
		if !reflect.DeepEqual(newDataIDs, oldDataIDs) {
			return waitIndex + 1, nil
		}
	}
}

func (c *Client) Close() {}
