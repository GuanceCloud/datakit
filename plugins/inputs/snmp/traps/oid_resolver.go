// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmprefiles"
	"gopkg.in/yaml.v2"
)

const defaultTrapDBFileNamePrefix string = "default_traps_db"

var nodesOIDThatShouldNeverMatch = []string{
	"1.3.6.1.4.1", // "iso.org.dod.internet.private.enterprises". This OID and all its parents are known "intermediate" nodes
	"1.3.6.1.4",   // "iso.org.dod.internet.private"
	"1.3.6.1",     // "iso.org.dod.internet"
	"1.3.6",       // "iso.org.dod"
	"1.3",         // "iso.org"
	"1",           // "iso"
}

type unmarshaller func(data []byte, v interface{}) error

// OIDResolver is a interface to get Trap and Variable metadata from OIDs.
type OIDResolver interface {
	GetTrapMetadata(trapOID string) (TrapMetadata, error)
	GetVariableMetadata(trapOID string, varOID string) (VariableMetadata, error)
}

// MultiFilesOIDResolver is an OIDResolver implementation that can be configured with multiple input files.
// Trap OIDs conflicts are resolved using the name of the source file in alphabetical order and by giving
// the less priority to Datadog's own database shipped with the agent.
// Variable OIDs conflicts are fully resolved by also looking at the trap OID. A given trap OID only
// exist in a single file (after the previous conflict resolution), meaning that we get the variable
// metadata from that same file.
type MultiFilesOIDResolver struct {
	traps TrapSpec
}

// NewMultiFilesOIDResolver creates a new MultiFilesOIDResolver instance by loading json or yaml files
// (optionnally gzipped) located in the directory conf.d/snmp/traps_db/.
func NewMultiFilesOIDResolver() (*MultiFilesOIDResolver, error) {
	oidResolver := &MultiFilesOIDResolver{traps: make(TrapSpec)}
	trapsDBRoot := snmprefiles.GetTrapsDBRoot()
	files, err := os.ReadDir(trapsDBRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir `%s`: %w", trapsDBRoot, err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("dir `%s` does not contain any trap db file", trapsDBRoot)
	}
	fileNames := getSortedFileNames(files)
	for _, fileName := range fileNames {
		err := oidResolver.updateFromFile(filepath.Join(trapsDBRoot, fileName))
		if err != nil {
			l.Warnf("unable to load trap db file %s: %s", fileName, err)
		}
	}
	return oidResolver, nil
}

// GetTrapMetadata returns TrapMetadata for a given trapOID.
func (or *MultiFilesOIDResolver) GetTrapMetadata(trapOID string) (TrapMetadata, error) {
	trapOID = strings.TrimSuffix(NormalizeOID(trapOID), ".0")
	trapData, ok := or.traps[trapOID]
	if !ok {
		return TrapMetadata{}, fmt.Errorf("trap OID %s is not defined", trapOID)
	}
	return trapData, nil
}

// GetVariableMetadata returns VariableMetadata for a given variableOID and trapOID.
// The trapOID should not be needed in theory but the Datadog Agent allows to define multiple variable names for the
// same OID as long as they are defined in different file. The trapOID is used to differentiate between these files.
func (or *MultiFilesOIDResolver) GetVariableMetadata(trapOID string, varOID string) (VariableMetadata, error) {
	trapOID = strings.TrimSuffix(NormalizeOID(trapOID), ".0")
	varOID = strings.TrimSuffix(NormalizeOID(varOID), ".0")
	trapData, ok := or.traps[trapOID]
	if !ok {
		return VariableMetadata{}, fmt.Errorf("trap OID %s is not defined", trapOID)
	}

	recreatedVarOID := varOID
	for {
		varData, ok := trapData.variableSpecPtr[recreatedVarOID]
		if ok {
			if varData.isIntermediateNode {
				// Found a known Node while climibing up the tree, no chance of finding a match higher
				return VariableMetadata{}, fmt.Errorf("variable OID %s is not defined", varOID)
			}
			return varData, nil
		}
		// No match for the current varOID, climb up the tree and retry
		lastDot := strings.LastIndex(recreatedVarOID, ".")
		if lastDot == -1 {
			break
		}
		recreatedVarOID = varOID[:lastDot]
	}
	return VariableMetadata{}, fmt.Errorf("variable OID %s is not defined", varOID)
}

func getSortedFileNames(files []fs.DirEntry) []string {
	if len(files) == 0 {
		return []string{}
	}
	// There should usually be one file provided by default and zero or more provided by the user
	userProvidedFileNames := make([]string, 0, len(files)-1)
	// Using a slice for error-proofing but there will usually be only one snmp provided file.
	defaultProvidedFileNames := make([]string, 0, 1)
	for _, file := range files {
		if file.IsDir() {
			l.Debugf("not loading traps data from path %s: file is directory", file.Name())
			continue
		}
		fileName := file.Name()
		if strings.HasPrefix(fileName, defaultTrapDBFileNamePrefix) {
			defaultProvidedFileNames = append(defaultProvidedFileNames, fileName)
		} else {
			userProvidedFileNames = append(userProvidedFileNames, file.Name())
		}
	}

	sort.Slice(userProvidedFileNames, func(i, j int) bool {
		return strings.ToLower(userProvidedFileNames[i]) < strings.ToLower(userProvidedFileNames[j])
	})
	sort.Slice(defaultProvidedFileNames, func(i, j int) bool {
		return strings.ToLower(defaultProvidedFileNames[i]) < strings.ToLower(defaultProvidedFileNames[j])
	})

	return append(defaultProvidedFileNames, userProvidedFileNames...)
}

func (or *MultiFilesOIDResolver) updateFromFile(fPath string) error {
	var fileReader io.ReadCloser
	fileReader, err := os.Open(filepath.Clean(fPath))
	if err != nil {
		return err
	}
	defer fileReader.Close() //nolint:errcheck
	if strings.HasSuffix(fPath, ".gz") {
		fPath = strings.TrimSuffix(fPath, ".gz")
		uncompressor, err := gzip.NewReader(fileReader)
		if err != nil {
			return fmt.Errorf("unable to uncompress gzip file %s", fPath)
		}
		defer uncompressor.Close() //nolint:errcheck
		fileReader = uncompressor
	}
	var unmarshalMethod unmarshaller = yaml.Unmarshal
	if strings.HasSuffix(fPath, ".json") {
		unmarshalMethod = json.Unmarshal
	}
	return or.updateFromReader(fileReader, unmarshalMethod)
}

func (or *MultiFilesOIDResolver) updateFromReader(reader io.Reader, unmarshalMethod unmarshaller) error {
	fileContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	var trapData trapDBFileContent
	err = unmarshalMethod(fileContent, &trapData)
	if err != nil {
		return err
	}

	or.updateResolverWithData(trapData)
	return nil
}

func (or *MultiFilesOIDResolver) updateResolverWithData(trapDB trapDBFileContent) {
	definedVariables := variableSpec{}

	allOIDs := make([]string, 0, len(trapDB.Variables))
	for variableOID := range trapDB.Variables {
		if !IsValidOID(variableOID) {
			l.Warnf("trap variable OID %s does not look like a valid OID", variableOID)
			continue
		}
		allOIDs = append(allOIDs, NormalizeOID(variableOID))
	}

	// "Fast" algorithm used to mark OID that act both as a variable and as a parent of other variable
	// with 'isNode: true'. i.e if an OID <FOO>.<BAR> exists in the trapsDB but <FOO> also exists in the trapsDB
	// then <FOO> acts as a 'Node' of the OID tree and should not be considered a match for resolving variables.
	// In this fast algorithm the list is sorted then each OID is compared with its successor. It the successor starts
	// with the current OID + a dot, then the current OID is a Node. 'Dots' are before digits in the lexicographic order.
	// Note that in practice, OIDs that act both as Node and Leaf of the OID tree is extremely rare and is not expected
	// in normal circumstamces. Thing is they sometimes exist.
	sort.Strings(allOIDs)
	for idx, variableOID := range allOIDs {
		isIntermediateNode := false
		if idx+1 < len(allOIDs) {
			nextOID := allOIDs[idx+1]
			isIntermediateNode = strings.HasPrefix(nextOID, variableOID+".")
		}

		variableData := trapDB.Variables[variableOID]
		variableData.isIntermediateNode = isIntermediateNode
		definedVariables[variableOID] = variableData
	}

	for _, nodeOID := range nodesOIDThatShouldNeverMatch {
		definedVariables[nodeOID] = VariableMetadata{Name: "unknown", isIntermediateNode: true}
	}

	for trapOID, trapData := range trapDB.Traps {
		if !IsValidOID(trapOID) {
			l.Errorf("trap OID %s does not look like a valid OID", trapOID)
			continue
		}
		trapOID := NormalizeOID(trapOID)
		if _, trapConflict := or.traps[trapOID]; trapConflict {
			l.Debugf("a trap with OID %s is defined in multiple traps db files", trapOID)
		}
		or.traps[trapOID] = TrapMetadata{
			Name:            trapData.Name,
			Description:     trapData.Description,
			MIBName:         trapData.MIBName,
			variableSpecPtr: definedVariables,
		}
	}
}
