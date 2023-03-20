// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmputil"
)

type deviceInfo struct {
	Ipt                   *Input
	IP                    string
	Namespace             string
	Subnet                string
	Session               snmputil.Session
	AutodetectProfile     bool
	CollectDeviceMetadata bool

	//--------------------------------------------------------------------------
	// those big stuff should pass from Input,
	// because they could be changed inside deviceInfo.
	Profile     string
	ProfileTags []string
	OidConfig   snmputil.OidConfig
	ProfileDef  *snmputil.ProfileDefinition
	Metadata    snmputil.MetadataConfig
	Metrics     []snmputil.MetricsConfig   // SNMP metrics definition
	MetricTags  []snmputil.MetricTagConfig // SNMP metric tags definition
	//--------------------------------------------------------------------------
}

// NewDeviceInfo why not use a struct to pass parameters? Because every parameter matters.
func NewDeviceInfo(
	ipt *Input,
	ip, namespace, subnet string,
	session snmputil.Session,
) *deviceInfo {
	return &deviceInfo{
		Ipt:                   ipt,
		IP:                    ip,
		Namespace:             namespace,
		Subnet:                subnet,
		Session:               session,
		AutodetectProfile:     ipt.autodetectProfile,
		CollectDeviceMetadata: true,
		Profile:               ipt.Profile,
		ProfileTags:           snmputil.CopyStrings(ipt.ProfileTags),
		OidConfig:             ipt.OidConfig.Copy(),
		ProfileDef:            ipt.ProfileDef.Copy(),
		Metadata:              snmputil.CopyMapStringMetadataResourceConfig(ipt.Metadata),
		Metrics:               snmputil.CopyMetricsConfigs(ipt.Metrics),
		MetricTags:            snmputil.CopyMetricTagConfigs(ipt.MetricTags),
	}
}

func (di *deviceInfo) initialize() error {
	if len(di.Profile) > 0 {
		err := di.refreshWithProfile(di.Profile)
		if err != nil {
			return fmt.Errorf("failed to refresh with profile `%s`: %w", di.Profile, err)
		}
	}
	return nil
}

//------------------------------------------------------------------------------

// bool: reachable    []string: tags    *snmputil.ResultValueStore: values    bool: isErrClosed
func (di *deviceInfo) getValuesAndTags() (bool, []string, *snmputil.ResultValueStore, error, bool) {
	var deviceReachable bool
	var checkErrors, tags []string

	err := di.Session.Connect()
	if err != nil {
		l.Warnf("Connect failed: err = (%v), ip = (%s)", err, di.IP)
		return false, nil, nil, err, false
	}
	defer func() {
		err := di.Session.Close()
		if err != nil {
			l.Warnf("failed to close session: err = (%v), ip = (%s)", err, di.IP)
		}
	}()

	// Check if the device is reachable
	getNextValue, err := di.Session.GetNext([]string{snmputil.DeviceReachableGetNextOid})
	if err != nil {
		deviceReachable = false
		checkErrors = append(checkErrors, fmt.Sprintf("check device reachable: failed: %v", err))
	} else {
		deviceReachable = true
		l.Debugf("check device reachable: success: %v", snmputil.PacketAsString(getNextValue))
	}

	err = di.doAutodetectProfile()
	if err != nil {
		checkErrors = append(checkErrors, fmt.Sprintf("failed to autodetect profile: %v", err))
	}

	tags = append(tags, di.ProfileTags...)

	valuesStore, err := snmputil.Fetch(di.Session, &snmputil.FetchOpts{
		OidConfig:          di.OidConfig,
		OidBatchSize:       defaultOidBatchSize,
		BulkMaxRepetitions: defaultBulkMaxRepetitions,
	})
	// l.Debugf("fetched values: %v", snmputil.ResultValueStoreAsString(valuesStore))

	if err != nil {
		checkErrors = append(checkErrors, fmt.Sprintf("failed to fetch values: %v", err))
	} else {
		tags = append(tags, snmputil.GetCheckInstanceMetricTags(di.MetricTags, valuesStore)...)
	}

	var joinedError error
	if len(checkErrors) > 0 {
		joinedError = errors.New(strings.Join(checkErrors, "; "))
	}

	for _, errStr := range checkErrors {
		if errStr == net.ErrClosed.Error() {
			return deviceReachable, tags, valuesStore, joinedError, true
		}
	}

	return deviceReachable, tags, valuesStore, joinedError, false
}

// refreshWithProfile refreshes config based on profile.
func (di *deviceInfo) refreshWithProfile(profile string) error {
	if _, ok := di.Ipt.Profiles[profile]; !ok {
		return fmt.Errorf("unknown profile `%s`", profile)
	}
	l.Debugf("Refreshing with profile `%s`", profile)
	tags := []string{"snmp_profile:" + profile}
	definition := di.Ipt.Profiles[profile]
	di.ProfileDef = &definition
	di.Profile = profile

	di.Metadata = snmputil.UpdateMetadataDefinitionWithLegacyFallback(definition.Metadata)
	di.Metrics = append(di.Metrics, definition.Metrics...)
	di.MetricTags = append(di.MetricTags, definition.MetricTags...)

	di.OidConfig.Clean()
	di.OidConfig.AddScalarOids(snmputil.ParseScalarOids(di.Metrics, di.MetricTags, di.Metadata, di.CollectDeviceMetadata))
	di.OidConfig.AddColumnOids(snmputil.ParseColumnOids(di.Metrics, di.Metadata, di.CollectDeviceMetadata))

	if definition.Device.Vendor != "" {
		tags = append(tags, "device_vendor:"+definition.Device.Vendor)
	}
	tags = append(tags, definition.StaticTags...)
	di.ProfileTags = tags
	return nil
}

func (di *deviceInfo) doAutodetectProfile() error {
	// Try to detect profile using device sysobjectid
	if di.AutodetectProfile {
		sysObjectID, err := snmputil.FetchSysObjectID(di.Session)
		if err != nil {
			return fmt.Errorf("failed to fetch sysobjectid: %w", err)
		}
		di.AutodetectProfile = false // do not try to auto detect profile next time

		profile, err := snmputil.GetProfileForSysObjectID(di.Ipt.Profiles, sysObjectID)
		if err != nil {
			return fmt.Errorf("failed to get profile sys object id for `%s`: %w", sysObjectID, err)
		}
		err = di.refreshWithProfile(profile)
		if err != nil {
			// Should not happen since the profile is one of those we matched in GetProfileForSysObjectID
			return fmt.Errorf("failed to refresh with profile `%s` detected using sysObjectID `%s`: %w", profile, sysObjectID, err)
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type deviceMetaData struct {
	collectMeta bool // collect meta if collect object.
	data        []string
}

func (dmd *deviceMetaData) Add(bys []byte) {
	dmd.data = append(dmd.data, string(bys))
}

// nolint:lll
// ReportNetworkDeviceMetadata reports device metadata.
func (di *deviceInfo) ReportNetworkDeviceMetadata(store *snmputil.ResultValueStore, origTags []string, metadataConfigs snmputil.MetadataConfig, collectTime time.Time, deviceStatus snmputil.DeviceStatus, outData *deviceMetaData) {
	tags := snmputil.CopyStrings(origTags)
	tags = snmputil.SortUniqInPlace(tags)

	metadataStore := snmputil.BuildMetadataStore(metadataConfigs, store)

	deviceID := di.getDeviceID()
	deviceIDTags := snmputil.SortUniqInPlace(di.getDeviceIDTags())
	tags = append(tags, deviceIDTags...)

	device := di.buildNetworkDeviceMetadata(deviceID, deviceIDTags, metadataStore, tags, deviceStatus)

	interfaces := snmputil.BuildNetworkInterfacesMetadata(deviceID, metadataStore)

	metadataPayloads := snmputil.BatchPayloads(di.Namespace, di.Subnet, collectTime, snmputil.PayloadMetadataBatchSize, device, interfaces)

	for _, payload := range metadataPayloads {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			l.Errorf("Error marshaling device metadata: %v", err)
			return
		}
		outData.Add(payloadBytes)
	}
}

// nolint:lll
func (di *deviceInfo) buildNetworkDeviceMetadata(deviceID string, idTags []string, store *snmputil.Store, tags []string, deviceStatus snmputil.DeviceStatus) snmputil.DeviceMetadata {
	var vendor, sysName, sysDescr, sysObjectID, location, serialNumber, version, productName, model, osName, osVersion, osHostname string
	if store != nil {
		sysName = store.GetScalarAsString("device.name")
		sysDescr = store.GetScalarAsString("device.description")
		sysObjectID = store.GetScalarAsString("device.sys_object_id")
		vendor = store.GetScalarAsString("device.vendor")
		location = store.GetScalarAsString("device.location")
		serialNumber = store.GetScalarAsString("device.serial_number")
		version = store.GetScalarAsString("device.version")
		productName = store.GetScalarAsString("device.product_name")
		model = store.GetScalarAsString("device.model")
		osName = store.GetScalarAsString("device.os_name")
		osVersion = store.GetScalarAsString("device.os_version")
		osHostname = store.GetScalarAsString("device.os_hostname")
	}

	// fallback to Device.Vendor for backward compatibility
	if di.ProfileDef != nil && vendor == "" {
		vendor = di.ProfileDef.Device.Vendor
	}

	return snmputil.DeviceMetadata{
		ID:           deviceID,
		IDTags:       idTags,
		Name:         sysName,
		Description:  sysDescr,
		IPAddress:    di.IP,
		SysObjectID:  sysObjectID,
		Location:     location,
		Profile:      di.Profile,
		Vendor:       vendor,
		Tags:         tags,
		Subnet:       di.Subnet,
		Status:       deviceStatus,
		SerialNumber: serialNumber,
		Version:      version,
		ProductName:  productName,
		Model:        model,
		OsName:       osName,
		OsVersion:    osVersion,
		OsHostname:   osHostname,
	}
}

func (di *deviceInfo) getDeviceID() string {
	return di.Namespace + ":" + di.IP
}

func (di *deviceInfo) getDeviceIDTags() []string {
	tags := []string{deviceNamespaceTagKey + ":" + di.Namespace, deviceIPTagKey + ":" + di.IP}
	sort.Strings(tags)
	return tags
}

//------------------------------------------------------------------------------
