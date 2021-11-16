package hostobject

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

var (
	HwCloud  = &hwcloud{}
	metadata = `{
  "random_seed": "z0f4IsSJHRKaKuyYkijZ9E+kxGtP+OxQLmDsMUwSklX5yxUUr+2v6759BZZIudkLRbCUa25I0Wk7Dr37fwF9mQUZ2GANc+KEEbgpSDRiYujlLF7AgNRFZbH9ES+C/0ZnM81nt5y1kc83amo3j/JFDLzYUvHCp2ZOYovBxokKk+AKZDZAVcYOLWGiq/+h1dN9+CevfCZSF82CT2afMO/BKwJkGr17Z6LYknu73ridlQgzRzVGGOg5NQi655xFn+eTdnFsEJ4HwBeAGdM9uiTHYGPDbf7QXCX2QKIPWyG/nAfwElkJ3i8ilgAbOhaN/YzKpx3l9STzDrYUAXJyxetZnie3e1g7WsN6nPh8CapRRwANJ/poxkYeiEHSzzu2zMJOq6E2IJ7TGcGSitBWSa8onNLcZ6yzYMvr+0pyPN3IRH5vcM4X/3IEM8nk5BfshwIxdKmQp+C72aGs7QJ4VeO6NxX4racee9fHAH67ObZE+0rObq2P+oxxO8fIAMDsm/0LVeqSJgBdd5j8WPgWEE9e6ak8viAvexChyJ/yBGgxmA3w6Ln8D/9sUaMPFelTgP2ZqQISQFmTqryxnk6q3675JOJPR6//+cqP71L0F8mqZDsWR+p8rIvBuob1Ore9ycPTJhpZdwIcTIvaYO2O7fHH8FTJCz7MGKOPYVex0JiJaWw=",
  "uuid": "f36e1d55-7d27-4772-9a3e-9737cae3b3a7",
  "availability_zone": "cn-east-3c",
  "enterprise_project_id": "0",
  "hostname": "ecs-b67f.novalocal",
  "launch_index": 0,
  "instance_type": "s6.small.1",
  "meta": {
    "metering.image_id": "6674d782-54ba-4f04-896d-95edd50f2eb9",
    "metering.imagetype": "gold",
    "metering.resourcespeccode": "s6.small.1.linux",
    "metering.cloudServiceType": "hws.service.type.ec2",
    "image_name": "CentOS 8.2 64bit",
    "os_bit": "64",
    "EcmResStatus": "",
    "cascaded.instance_extrainfo": "pcibridge:1",
    "metering.resourcetype": "1",
    "vpc_id": "868c3644-3850-4551-bd69-2465367be012",
    "os_type": "Linux",
    "charging_mode": "0",
    "__support_agent_list": "hss"
  },
  "region_id": "cn-east-3",
  "project_id": "09bb26eeef80f46d2fdec014d81fe726",
  "name": "ecs-b67f"
}`
	netWorkdata = `{
  "services": [
    {
      "type": "dns",
      "address": "100.125.1.250"
    },
    {
      "type": "dns",
      "address": "100.125.64.250"
    }
  ],
  "qos": {
    "instance_min_bandwidth": 100,
    "instance_max_bandwidth": 800
  },
  "networks": [
    {
      "network_id": "cde27763-fc56-4c8b-9068-a36fefae49da",
      "type": "ipv4_dhcp",
      "link": "tap7bdc4804-58",
      "id": "network0"
    }
  ],
  "links": [
    {
      "vif_id": "7bdc4804-58f0-4828-b8b0-15a43515c888",
      "public_ipv4": "121.37.180.221",
      "ethernet_mac_address": "fa:16:3e:eb:19:d2",
      "mtu": null,
      "local_ipv4": "192.168.0.199",
      "type": "cascading",
      "id": "tap7bdc4804-58"
    }
  ]
}`
	privateIP = `Sys-WebServer`
)

func testGetHwMeta() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openstack/latest/meta_data.json":
			fmt.Fprint(w, metadata)
		case "/openstack/latest/network_data.json":
			fmt.Fprint(w, netWorkdata)
		case "/latest/meta-data/security-groups":
			fmt.Fprint(w, privateIP)
		default:
			fmt.Fprintf(w, "bad response")
		}
	}))
	HwCloud.baseURL = ts.URL
	return ts
}

func parseHwcloudData() *hwMetaData {
	ts := testGetHwMeta()
	resp := metadataGet(HwCloud.baseURL + "/openstack/latest/meta_data.json")
	defer ts.Close()
	model := &hwMetaData{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		l.Errorf("marshal json failed: %s", err)
		return nil
	}
	return model
}

func parseHwcloudNetData() *netWorkData {
	ts := testGetHwMeta()
	resp := metadataGet(HwCloud.baseURL + "/openstack/latest/network_data.json")
	defer ts.Close()
	model := &netWorkData{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		l.Errorf("marshal json failed: %s", err)
		return nil
	}
	return model
}

func TestHwcloud_InstanceID(t *testing.T) {
	hwMetaData := parseHwcloudData()
	tu.Assert(t, hwMetaData.InstanceID == "f36e1d55-7d27-4772-9a3e-9737cae3b3a7", "Hwcloud_InstanceID")
}

func TestHwcloud_InstanceName(t *testing.T) {
	hwMetaData := parseHwcloudData()
	tu.Assert(t, hwMetaData.InstanceName == "ecs-b67f", "Hwcloud_InstanceName")
}

func TestHwcloud_InstanceNetworkType(t *testing.T) {
	hwMetaData := parseHwcloudNetData()
	tu.Assert(t, hwMetaData.NetWork[0].InstanceNetworkType == "ipv4_dhcp", "Hwcloud_InstanceNetworkType")
}

func TestHwcloud_InstanceType(t *testing.T) {
	hwMetaData := parseHwcloudData()
	tu.Assert(t, hwMetaData.InstanceType == "s6.small.1", "Hwcloud_InstanceType")
}

func TestHwcloud_Region(t *testing.T) {
	hwMetaData := parseHwcloudData()
	tu.Assert(t, hwMetaData.Region == "cn-east-3", "Hwcloud_Region")
}

func TestHwcloud_PrivateIP(t *testing.T) {
	hwMetaData := parseHwcloudNetData()
	tu.Assert(t, hwMetaData.Links[0].PrivateIP == "192.168.0.199", "Hwcloud_PrivateIP")
}

func TestHwcloud_ZoneID(t *testing.T) {
	hwMetaData := parseHwcloudData()
	tu.Assert(t, hwMetaData.ZoneID == "cn-east-3c", "Hwcloud_ZoneID")
}

func TestHwcloud_SecurityGroupID(t *testing.T) {
	ts := testGetHwMeta()
	resp := metaGet(ts.URL + "/latest/meta-data/security-groups")
	tu.Assert(t, resp == "Sys-WebServer", "Hwcloud_SecurityGroupID")
}

func TestHwcloud_InstanceChargeType(t *testing.T) {
	ts := testGetHwMeta()
	resp := metadataGet(HwCloud.baseURL + "/openstack/latest/meta_data.json")
	defer ts.Close()
	model := struct {
		HwcloudInstanceChargeType string
	}{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		t.Errorf("marshal json failed: %s", err)
	}
	tu.Assert(t, model.HwcloudInstanceChargeType == "", "Hwcloud_ChargeType")
}

func TestHwcloud_WrongRouter(t *testing.T) {
	ts := testGetHwMeta()
	resp := metadataGet(HwCloud.baseURL + "/openstack/latest/wrongCase")
	defer ts.Close()
	tu.Assert(t, string(resp) == "bad response", "Hwcloud_WrongRouter")
}
