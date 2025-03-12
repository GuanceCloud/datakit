// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
	"gopkg.in/yaml.v2"
)

func TestInput_doCollectUserObject(t *testing.T) {
	type args struct {
		deviceIP string
		sess     snmputil.Session
		ipt      *Input
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "collect_object_test",
			args: args{
				deviceIP: "1.2.3.4",
				sess:     getMockSession01(),
				ipt:      getMockInput01(),
			},
			want: []string{
				"snmp_net_device,ip=1.2.3.4,name=_ZY_WLC,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12 netUptime=204572431,system_location=\"Shenzhen China\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeder := dkio.NewMockedFeeder()

			ipt := tt.args.ipt
			ipt.feeder = feeder
			ipt.UserProfileStore = snmputil.NewUserProfileStore()

			di := &deviceInfo{UserProfileDefinition: snmputil.NewUserProfileDefinition()}
			err := yaml.Unmarshal([]byte(zabbixYaml), di.UserProfileDefinition)
			assert.NoError(t, err)
			di.Session = tt.args.sess
			di.UserProfileDefinition.Class = "net_device"
			di.Ipt = ipt
			di.preProcess()

			ipt.doCollectUserObject(tt.args.deviceIP, di)

			pts, err := feeder.AnyPoints(time.Second * 10)
			assert.NoError(t, err)

			arr := GetPointLineProtos(pts)
			t.Log(arr)
			assert.Equal(t, tt.want, arr)
		})
	}
}

func TestInput_doCollectUserMetrics(t *testing.T) {
	type args struct {
		deviceIP string
		sess     snmputil.Session
		ipt      *Input
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "collect_metrics_test",
			args: args{
				deviceIP: "1.2.3.4",
				sess:     getMockSession01(),
				ipt:      getMockInput01(),
			},
			want: []string{
				"snmp_net_device,component=system,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.1.3.0,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12 netUptime=204572431",
				"snmp_net_device,component=cpu,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_name=SRU\\ Board\\ 0 cpuUsage=37",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.1,snmp_value=1,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ InLoopBack0\\ Interface,unit_desc=InLoopBack0,unit_name=InLoopBack0,unit_status=1,unit_type=24 ifHighSpeed=0",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.2,snmp_value=1,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ NULL0\\ Interface,unit_desc=NULL0,unit_name=NULL0,unit_status=1,unit_type=1 ifHighSpeed=0",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.3,snmp_value=1,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/1\\ Interface,unit_desc=GigabitEthernet0/0/1,unit_name=GigabitEthernet0/0/1,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.4,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/2\\ Interface,unit_desc=GigabitEthernet0/0/2,unit_name=GigabitEthernet0/0/2,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.5,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/3\\ Interface,unit_desc=GigabitEthernet0/0/3,unit_name=GigabitEthernet0/0/3,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.6,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/4\\ Interface,unit_desc=GigabitEthernet0/0/4,unit_name=GigabitEthernet0/0/4,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.7,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/5\\ Interface,unit_desc=GigabitEthernet0/0/5,unit_name=GigabitEthernet0/0/5,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.8,snmp_value=1,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/6\\ Interface,unit_desc=GigabitEthernet0/0/6,unit_name=GigabitEthernet0/0/6,unit_status=1,unit_type=6 ifHighSpeed=100",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.9,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/7\\ Interface,unit_desc=GigabitEthernet0/0/7,unit_name=GigabitEthernet0/0/7,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.10,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ GigabitEthernet0/0/8\\ Interface,unit_desc=GigabitEthernet0/0/8,unit_name=GigabitEthernet0/0/8,unit_status=1,unit_type=6 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.11,snmp_value=1,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ Vlanif1\\ Interface,unit_desc=Vlanif1,unit_name=Vlanif1,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.12,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ Vlanif305\\ Interface,unit_desc=Vlanif305,unit_name=Vlanif305,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.13,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=connect\\ ZY_DSW,unit_desc=Eth-Trunk1,unit_name=Eth-Trunk1,unit_status=1,unit_type=53 ifHighSpeed=0",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.196,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=wireless-zhuyun-2F,unit_desc=Vlanif302,unit_name=Vlanif302,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.197,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=wireless-guest,unit_desc=Vlanif300,unit_name=Vlanif300,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.198,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=wireless-zhuyun-4F,unit_desc=Vlanif304,unit_name=Vlanif304,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.199,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=wireless-zhuyun-1F,unit_desc=Vlanif301,unit_name=Vlanif301,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.200,snmp_value=2,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=wireless-zhuyun-3F,unit_desc=Vlanif303,unit_name=Vlanif303,unit_status=1,unit_type=53 ifHighSpeed=1000",
				"snmp_net_device,Application=Network\\ Interfaces,ip=1.2.3.4,name=_ZY_WLC,oid=1.3.6.1.2.1.31.1.1.1.15.201,snmp_value=1,sys_name=ZY_WLC,sys_object_id=.1.3.6.1.4.1.2011.2.240.12,unit_alias=HUAWEI\\,\\ AC\\ Series\\,\\ Ethernet0/0/47\\ Interface,unit_desc=Ethernet0/0/47,unit_name=Ethernet0/0/47,unit_status=1,unit_type=6 ifHighSpeed=1000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeder := dkio.NewMockedFeeder()

			ipt := tt.args.ipt
			ipt.feeder = feeder
			ipt.UserProfileStore = snmputil.NewUserProfileStore()

			di := &deviceInfo{UserProfileDefinition: snmputil.NewUserProfileDefinition()}
			err := yaml.Unmarshal([]byte(zabbixYaml), di.UserProfileDefinition)
			assert.NoError(t, err)
			di.Session = tt.args.sess
			di.UserProfileDefinition.Class = "net_device"
			di.Ipt = ipt
			di.preProcess()

			ipt.doCollectUserMetrics(tt.args.deviceIP, di)

			pts, err := feeder.AnyPoints(time.Second * 10)
			assert.NoError(t, err)

			arr := GetPointLineProtos(pts)
			assert.Equal(t, tt.want, arr)
		})
	}
}
