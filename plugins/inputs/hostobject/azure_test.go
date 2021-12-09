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
	Azure     = &azure{}
	Azuredata = `{
  "compute": {
    "azEnvironment": "AzurePublicCloud",
    "customData": "",
    "evictionPolicy": "",
    "isHostCompatibilityLayerVm": "false",
    "licenseType": "",
    "location": "eastus",
    "name": "azure-zy-test",
    "offer": "0001-com-ubuntu-server-focal",
    "osProfile": {
      "adminUsername": "azureuser",
      "computerName": "azure-zy-test",
      "disablePasswordAuthentication": "true"
    },
    "osType": "Linux",
    "placementGroupId": "",
    "plan": {
      "name": "",
      "product": "",
      "publisher": ""
    },
    "platformFaultDomain": "0",
    "platformUpdateDomain": "0",
    "priority": "",
    "provider": "Microsoft.Compute",
    "publicKeys": [
      {
        "keyData": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFGE9mH9mLp8ZZucyoCIhCrcvm\r\ndfM2fNU8/EMwRV1Ll7Q42G10rOuQgDsG3G31CSvI07f0+NXC5zDCIKsUyf04B13L\r\nCL+27J3ILaWRfIex+kUKUcr2XUQGaaC02DWIGl1HxrMlRiIlLZI2uRHp6yHlCQKE\r\nUgXsPg1/rQTYSTiVyYctbxmjjnE2+qYhJdDxt91PLfF+3uP8x5J0CT/1XW8MH8Br\r\n8SopW4ch+V5tMSK/uYMi+5oI0xTIVXd+HJqZfQNzEZtNX0wiSeoDtuj3VdnnaKz/\r\n27X+yWn7q77Vlf84DU3qjqliPloJINpLhu92j2LiiWs90ejDFMY787nsX4qsm1In\r\ng3ylLhHIOq/9SVCLakCmak5WhDsErBrcfGEPxrE+NNQVvuWHK4tN5Tes7BSZF6sP\r\n1GEjmL7nPMb16spVvcw76O1yvuKqgh9YM2F1xmP3nCrEayHGbnS/4qvxz9wRAAjH\r\nlQwRYw2sE0/gBk+uu0sckJdxbRr99daz7BI07L0= generated-by-azure\r\n",
        "path": "/home/azureuser/.ssh/authorized_keys"
      }
    ],
    "publisher": "canonical",
    "resourceGroupName": "zy-test",
    "resourceId": "/subscriptions/301375df-789c-46b5-a567-29da356bd735/resourceGroups/zy-test/providers/Microsoft.Compute/virtualMachines/azure-zy-test",
    "securityProfile": {
      "secureBootEnabled": "false",
      "virtualTpmEnabled": "false"
    },
    "sku": "20_04-lts-gen2",
    "storageProfile": {
      "dataDisks": [],
      "imageReference": {
        "id": "",
        "offer": "0001-com-ubuntu-server-focal",
        "publisher": "canonical",
        "sku": "20_04-lts-gen2",
        "version": "latest"
      },
      "osDisk": {
        "caching": "ReadWrite",
        "createOption": "FromImage",
        "diffDiskSettings": {
          "option": ""
        },
        "diskSizeGB": "30",
        "encryptionSettings": {
          "enabled": "false"
        },
        "image": {
          "uri": ""
        },
        "managedDisk": {
          "id": "/subscriptions/301375df-789c-46b5-a567-29da356bd735/resourceGroups/zy-test/providers/Microsoft.Compute/disks/azure-zy-test_OsDisk_1_ae0a9e99035941c1a352003e95e0a57d",
          "storageAccountType": "Premium_LRS"
        },
        "name": "azure-zy-test_OsDisk_1_ae0a9e99035941c1a352003e95e0a57d",
        "osType": "Linux",
        "vhd": {
          "uri": ""
        },
        "writeAcceleratorEnabled": "false"
      },
      "resourceDisk": {
        "size": "34816"
      }
    },
    "subscriptionId": "301375df-789c-46b5-a567-29da356bd735",
    "tags": "",
    "tagsList": [],
    "userData": "",
    "version": "20.04.202110260",
    "vmId": "fb17b79a-599e-4530-8713-283be4a5bdf4",
    "vmScaleSetName": "",
    "vmSize": "Standard_B1s",
    "zone": "1"
  },
  "network": {
    "interface": [
      {
        "ipv4": {
          "ipAddress": [
            {
              "privateIpAddress": "10.0.0.4",
              "publicIpAddress": ""
            }
          ],
          "subnet": [
            {
              "address": "10.0.0.0",
              "prefix": "24"
            }
          ]
        },
        "ipv6": {
          "ipAddress": []
        },
        "macAddress": "000D3A12A159"
      }
    ]
  }
}`
	PrivateIP = `10.0.0.4`
)

func testGetAzure() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metadata/instance":
			if r.URL.RawQuery != "api-version=2021-02-01" {
				fmt.Fprintf(w, "bad response")
				break
			}
			fmt.Fprint(w, Azuredata)
		case "/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress":
			if r.URL.RawQuery != "api-version=2021-02-01" {
				fmt.Fprintf(w, "bad response")
				break
			}
			fmt.Fprint(w, PrivateIP)
		default:
			fmt.Fprintf(w, "bad response")
		}
	}))
	Azure.baseURL = ts.URL
	return ts
}

func parseAzureData() *azureMetaData {
	ts := testGetAzure()
	resp := metadataGetByHeader(Azure.baseURL + "/metadata/instance?api-version=2021-02-01")
	defer ts.Close()
	model := &azureMetaData{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		l.Errorf("marshal json failed: %s", err)
		return nil
	}
	return model
}

func TestAzure_InstanceID(t *testing.T) {
	azureMetaData := parseAzureData()
	tu.Assert(t, azureMetaData.Compute.InstanceID == "fb17b79a-599e-4530-8713-283be4a5bdf4", "Azure_InstanceID")
}

func TestAzure_InstanceName(t *testing.T) {
	azureMetaData := parseAzureData()
	tu.Assert(t, azureMetaData.Compute.InstanceName == "azure-zy-test", "Azure_InstanceName")
}

func TestAzure_InstanceType(t *testing.T) {
	azureMetaData := parseAzureData()
	tu.Assert(t, azureMetaData.Compute.InstanceType == "Standard_B1s", "Azure_InstanceType")
}

func TestAzure_ZoneID(t *testing.T) {
	azureMetaData := parseAzureData()
	tu.Assert(t, azureMetaData.Compute.ZoneID == "1", "Azure_ZoneID")
}

func TestAzure_Region(t *testing.T) {
	azureMetaData := parseAzureData()
	tu.Assert(t, azureMetaData.Compute.Region == "eastus", "Azure_Region")
}

func TestAzure_PrivateIP(t *testing.T) {
	ts := testGetAzure()
	resp := metadataGetByHeader(Azure.baseURL + "/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2021-02-01")
	defer ts.Close()
	tu.Assert(t, string(resp) == "10.0.0.4", "Azure_PrivateIP")
}

func TestAzure_InstanceChargeType(t *testing.T) {
	ts := testGetAzure()
	resp := metadataGetByHeader(Azure.baseURL + "/metadata/instance?api-version=2021-02-01")
	defer ts.Close()
	model := struct {
		AzureInstanceChargeType string
	}{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		t.Errorf("marshal json failed: %s", err)
	}
	tu.Assert(t, model.AzureInstanceChargeType == "", "Azure_InstanceChargeType")
}

func TestAzure_WrongRouter(t *testing.T) {
	ts := testGetAzure()
	resp := metadataGetByHeader(Azure.baseURL + "/metadata/instance?api-version=2021-02-01/wrongCase")
	defer ts.Close()
	tu.Assert(t, string(resp) == "bad response", "Azure_WrongRouter")
}
