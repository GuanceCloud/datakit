package nacos

import (
	"strings"
	"time"

	"github.com/GuanceCloud/confd/log"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"fmt"
	"net/url"
	"strconv"
)

type Client struct {
	// "serverConfigs" + "clientConfig"
	Configs config_client.IConfigClient
	// Listener List. example: map["namespace_name/group_name/dataID_name]struct{})
	Listeners map[string]struct{}
	// cycle time interval, second
	NacosInterval int
	ExitWatchCh   chan error
}

// NewNacosClient create nacod client
func NewNacosClient(
	backendNodes []string,
	password,
	username,
	namespace,
	accessKey,
	secretKey string,
	nacosInterval int,
) (c *Client, err error) {

	serverConfigs := []constant.ServerConfig{}
	for _, backendNode := range backendNodes {
		nacosUrl, err := url.Parse(backendNode) // like "http://127.0.0.1:8848"
		if err != nil {
			return nil, fmt.Errorf("nacos backendNode url Parse : %v", err)
		}

		port, err := strconv.Atoi(nacosUrl.Port())
		if err != nil {
			return nil, fmt.Errorf("nacos backendNode port Atoi : %v", err)
		}

		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: nacosUrl.Hostname(),
			Port:   uint64(port),
		})
	}

	clientConfig := constant.ClientConfig{
		Endpoint:       backendNodes[0],
		NamespaceId:    namespace,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		TimeoutMs:      10 * 1000, // HTTP request timeout, in milliseconds
		ListenInterval: 30 * 1000, // Listening interval, in milliseconds (only valid in ConfigClient)
		BeatInterval:   5 * 1000,  // Heartbeat interval, in milliseconds (only valid in ServiceClient)
	}

	// config
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("nacos clients.CreateConfigClient : %v", err)
	}

	if nacosInterval < 1 {
		nacosInterval = 60
	}

	c = &Client{
		Configs:       configClient,
		Listeners:     make(map[string]struct{}),
		NacosInterval: nacosInterval,
		ExitWatchCh:   make(chan error, 1),
	}

	return
}

// GetValues get all dataID in this namespace
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("nacos GetValues keys is nil")
	}

	// SearchConfig
	content, err := c.Configs.SearchConfig(vo.SearchConfigParam{PageSize: 65535, Search: "blur"})
	if err != nil {
		return nil, fmt.Errorf("nacos configClient.SearchConfig : %v", err)
	}

	// to map[string]string
	values := map[string]string{}
	for _, pageItem := range content.PageItems {
		key := keys[0] + "/" + pageItem.Group + "/" + pageItem.DataId
		values[key] = pageItem.Content
	}

	return values, nil
}

// WatchPrefix watch all namespace
// if creat new dataID, perhaps delay 60S
// @prefix @keys will no useful
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {

	// list of all dataIDs in this namespace, the key like "/group_name/dataID_name"
	watchDataIDs, err := c.GetValues(keys)
	if err != nil {
		return waitIndex + 1, fmt.Errorf("nacos WatchPrefix : %v", err)
	}

	// cancel some listener
	for key, _ := range c.Listeners {
		// check if already linen
		if _, ok := watchDataIDs[key]; ok {
			continue
		}

		// group dataID at the tail
		strs := strings.Split(key, "/")
		length := len(strs)
		if length < 3 {
			continue
		}
		group := strs[length-2]
		dataID := strs[length-1]

		err = c.Configs.CancelListenConfig(vo.ConfigParam{
			DataId: dataID,
			Group:  group,
		})
		if err != nil {
			log.Error("cancel Watch error: %s", err.Error())
		}

		// delete on Listeners
		delete(c.Listeners, key)
	}

	// creat listeners
	for key, _ := range watchDataIDs {
		// check if already linen
		if _, ok := c.Listeners[key]; ok {
			continue
		}

		// group dataID at the tail
		strs := strings.Split(key, "/")
		length := len(strs)
		if length < 3 {
			continue
		}
		group := strs[length-2]
		dataID := strs[length-1]

		err = c.Configs.ListenConfig(vo.ConfigParam{
			DataId: dataID,
			Group:  group,
			OnChange: func(namespace, group, dataId, data string) {
				// exit watch
				c.ExitWatchCh <- nil
			},
		})
		if err != nil {
			log.Error("Watch error: %s", err.Error())
		}

		// add to Listeners
		c.Listeners[key] = struct{}{}
	}

	// cycle time interval
	tick := time.NewTicker(time.Second * time.Duration(c.NacosInterval))
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
		case <-stopChan:
			c.cancelWatches()
			return waitIndex, fmt.Errorf("stopChan")
		case err := <-c.ExitWatchCh:
			return waitIndex + 1, err
		}

		// cycle get data to find new dataID
		newDataIDs, err := c.GetValues(keys)
		if err != nil {
			return waitIndex + 1, fmt.Errorf("nacos WatchPrefix : %v", err)
		}
		for key, _ := range newDataIDs {
			if _, ok := c.Listeners[key]; !ok {
				// find new dataID
				return waitIndex + 1, nil
			}
		}
	}
}

func (c *Client) cancelWatches() {}
