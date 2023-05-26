// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

//------------------------------------------------------------------------------

func fetchColumnOidsWithBatching(sess Session, oids map[string]string, oidBatchSize int, bulkMaxRepetitions uint32, fetchStrategy columnFetchStrategy) (ColumnResultValuesType, error) { //nolint:lll
	retValues := make(ColumnResultValuesType, len(oids))

	columnOids := getOidsMapKeys(oids)
	sort.Strings(columnOids) // sorting ColumnOids to make them deterministic for testing purpose
	batches, err := CreateStringBatches(columnOids, oidBatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create column oid batches: %w", err)
	}

	for _, batchColumnOids := range batches {
		oidsToFetch := make(map[string]string, len(batchColumnOids))
		for _, oid := range batchColumnOids {
			oidsToFetch[oid] = oids[oid]
		}

		results, err := fetchColumnOids(sess, oidsToFetch, bulkMaxRepetitions, fetchStrategy)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch column oids: %w", err)
		}

		for columnOid, instanceOids := range results {
			if _, ok := retValues[columnOid]; !ok {
				retValues[columnOid] = instanceOids
				continue
			}
			for oid, value := range instanceOids {
				retValues[columnOid][oid] = value
			}
		}
	}
	return retValues, nil
}

// fetchColumnOids has an `oids` argument representing a `map[string]string`,
// the key of the map is the column oid, and the value is the oid used to fetch the next value for the column.
// The value oid might be equal to column oid or a row oid of the same column.
func fetchColumnOids(sess Session, oids map[string]string, bulkMaxRepetitions uint32, fetchStrategy columnFetchStrategy) (ColumnResultValuesType, error) { //nolint:lll
	returnValues := make(ColumnResultValuesType, len(oids))
	alreadyProcessedOids := make(map[string]bool)
	curOids := oids
	for {
		if len(curOids) == 0 {
			break
		}
		// l.Debugf("fetch column: request oids (maxRep:%d,fetchStrategy:%s): %v", bulkMaxRepetitions, fetchStrategy, curOids)
		var columnOids, requestOids []string
		for k, v := range curOids {
			if alreadyProcessedOids[v] {
				l.Debugf("fetch column: OID already processed: %s", v)
				continue
			}
			alreadyProcessedOids[v] = true
			columnOids = append(columnOids, k)
			requestOids = append(requestOids, v)
		}
		if len(columnOids) == 0 {
			break
		}
		// sorting ColumnOids and requestOids to make them deterministic for testing purpose
		sort.Strings(columnOids)
		sort.Strings(requestOids)

		results, err := getResults(sess, requestOids, bulkMaxRepetitions, fetchStrategy)
		if err != nil {
			return nil, err
		}
		newValues, nextOids := ResultToColumnValues(columnOids, results)
		updateColumnResultValues(returnValues, newValues)
		curOids = nextOids
	}
	return returnValues, nil
}

func getResults(sess Session, requestOids []string, bulkMaxRepetitions uint32, fetchStrategy columnFetchStrategy) (*gosnmp.SnmpPacket, error) {
	var results *gosnmp.SnmpPacket
	if sess.GetVersion() == gosnmp.Version1 || fetchStrategy == useGetNext {
		// snmp v1 doesn't support GetBulk
		getNextResults, err := sess.GetNext(requestOids)
		if err != nil {
			l.Debugf("fetch column: failed getting oids `%v` using GetNext: %s", requestOids, err)
			return nil, fmt.Errorf("fetch column: failed getting oids `%v` using GetNext: %w", requestOids, err)
		}
		results = getNextResults
		// l.Debugf("fetch column: GetNext results: %v", PacketAsString(results))
	} else {
		getBulkResults, err := sess.GetBulk(requestOids, bulkMaxRepetitions)
		if err != nil {
			l.Debugf("fetch column: failed getting oids `%v` using GetBulk: %s", requestOids, err)
			return nil, fmt.Errorf("fetch column: failed getting oids `%v` using GetBulk: %w", requestOids, err)
		}
		results = getBulkResults
		// l.Debugf("fetch column: GetBulk results: %v", PacketAsString(results))
	}
	return results, nil
}

func updateColumnResultValues(valuesToUpdate ColumnResultValuesType, extraValues ColumnResultValuesType) {
	for columnOid, columnValues := range extraValues {
		for oid, value := range columnValues {
			if _, ok := valuesToUpdate[columnOid]; !ok {
				valuesToUpdate[columnOid] = make(map[string]ResultValue)
			}
			valuesToUpdate[columnOid][oid] = value
		}
	}
}

func getOidsMapKeys(oidsMap map[string]string) []string {
	keys := make([]string, len(oidsMap))
	i := 0
	for k := range oidsMap {
		keys[i] = k
		i++
	}
	return keys
}

//------------------------------------------------------------------------------

func fetchScalarOidsWithBatching(sess Session, oids []string, oidBatchSize int) (ScalarResultValuesType, error) {
	retValues := make(ScalarResultValuesType, len(oids))

	batches, err := CreateStringBatches(oids, oidBatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create oid batches: %w", err)
	}

	for _, batchOids := range batches {
		results, err := fetchScalarOids(sess, batchOids)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch scalar oids: %w", err)
		}
		for k, v := range results {
			retValues[k] = v
		}
	}
	return retValues, nil
}

func fetchScalarOids(sess Session, oids []string) (ScalarResultValuesType, error) {
	packet, err := doFetchScalarOids(sess, oids)
	if err != nil {
		return nil, err
	}
	values := ResultToScalarValues(packet)
	retryFailedScalarOids(sess, packet, values)
	return values, nil
}

// retryFailedScalarOids retries on NoSuchObject or NoSuchInstance for scalar oids not ending with `.0`.
// This helps keeping compatibility with python implementation.
// This is not need in normal circumstances where scalar OIDs end with `.0`.
// If the oid does not end with `.0`, we will retry by appending `.0` to it.
func retryFailedScalarOids(sess Session, results *gosnmp.SnmpPacket, valuesToUpdate ScalarResultValuesType) {
	retryOids := make(map[string]string)
	for _, variable := range results.Variables {
		oid := strings.TrimLeft(variable.Name, ".")
		if (variable.Type == gosnmp.NoSuchObject || variable.Type == gosnmp.NoSuchInstance) && !strings.HasSuffix(oid, ".0") {
			retryOids[oid] = oid + ".0"
		}
	}
	if len(retryOids) > 0 {
		fetchOids := make([]string, 0, len(retryOids))
		for _, oid := range retryOids {
			fetchOids = append(fetchOids, oid)
		}
		sort.Strings(fetchOids) // needed for stable tests since fetchOids order (from a map values) is undefined
		retryResults, err := doFetchScalarOids(sess, fetchOids)
		if err != nil {
			l.Debugf("failed to oids `%v` on retry: %v", retryOids, err)
		} else {
			retryValues := ResultToScalarValues(retryResults)
			for initialOid, actualOid := range retryOids {
				if value, ok := retryValues[actualOid]; ok {
					valuesToUpdate[initialOid] = value
				}
			}
		}
	}
}

func doFetchScalarOids(session Session, oids []string) (*gosnmp.SnmpPacket, error) {
	var results *gosnmp.SnmpPacket
	if session.GetVersion() == gosnmp.Version1 {
		// When using snmp v1, if one of the oids return a NoSuchName, all oids will have value of Null.
		// The response will contain Error=NoSuchName and ErrorIndex with index of the erroneous oid.
		// If that happen, we remove the erroneous oid and try again until we succeed or until there is no oid anymore.
		for {
			scalarOids, err := doDoFetchScalarOids(session, oids)
			if err != nil {
				return nil, err
			}
			if scalarOids.Error == gosnmp.NoSuchName {
				zeroBaseIndex := int(scalarOids.ErrorIndex) - 1 // ScalarOids.ErrorIndex is 1-based
				if (zeroBaseIndex < 0) || (zeroBaseIndex > len(oids)-1) {
					return nil, fmt.Errorf("invalid ErrorIndex `%d` when fetching oids `%v`", scalarOids.ErrorIndex, oids)
				}
				oids = append(oids[:zeroBaseIndex], oids[zeroBaseIndex+1:]...)
				if len(oids) == 0 {
					// If all oids are not found, return an empty packet with no variable and no error
					return &gosnmp.SnmpPacket{}, nil
				}
				continue
			}
			results = scalarOids
			break
		}
	} else {
		scalarOids, err := doDoFetchScalarOids(session, oids)
		if err != nil {
			return nil, err
		}
		results = scalarOids
	}
	return results, nil
}

func doDoFetchScalarOids(session Session, oids []string) (*gosnmp.SnmpPacket, error) {
	// l.Debugf("fetch scalar: request oids: %v", oids)
	results, err := session.Get(oids)
	if err != nil {
		l.Debugf("fetch scalar: error getting oids `%v`: %v", oids, err)
		return nil, fmt.Errorf("fetch scalar: error getting oids `%v`: %w", oids, err)
	}
	// l.Debugf("fetch scalar: results: %s", PacketAsString(results))
	return results, nil
}

//------------------------------------------------------------------------------

type columnFetchStrategy int

const (
	useGetBulk columnFetchStrategy = iota
	useGetNext
)

func (c columnFetchStrategy) String() string {
	switch c {
	case useGetBulk:
		return "useGetBulk"
	case useGetNext:
		return "useGetNext"
	default:
		return strconv.Itoa(int(c))
	}
}

type FetchOpts struct {
	OidConfig          OidConfig
	OidBatchSize       int
	BulkMaxRepetitions uint32
}

// Fetch oid values from device.
func Fetch(sess Session, config *FetchOpts) (*ResultValueStore, error) {
	// fetch scalar values
	scalarResults, err := fetchScalarOidsWithBatching(sess, config.OidConfig.ScalarOids, config.OidBatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch scalar oids with batching: %v", err) //nolint:errorlint
	}

	// fetch column values
	oids := make(map[string]string, len(config.OidConfig.ColumnOids))
	for _, value := range config.OidConfig.ColumnOids {
		oids[value] = value
	}

	columnResults, err := fetchColumnOidsWithBatching(sess, oids, config.OidBatchSize, config.BulkMaxRepetitions, useGetBulk)
	if err != nil {
		l.Debugf("failed to fetch oids with GetBulk batching: %v", err)

		columnResults, err = fetchColumnOidsWithBatching(sess, oids, config.OidBatchSize, config.BulkMaxRepetitions, useGetNext)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch oids with GetNext batching: %v", err) //nolint:errorlint
		}
	}

	return &ResultValueStore{ScalarValues: scalarResults, ColumnValues: columnResults}, nil
}

//------------------------------------------------------------------------------
