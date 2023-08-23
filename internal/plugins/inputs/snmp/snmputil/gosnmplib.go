// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

//------------------------------------------------------------------------------

// GetAuthProtocol converts auth protocol from string to type.
func GetAuthProtocol(authProtocolStr string) (gosnmp.SnmpV3AuthProtocol, error) {
	var authProtocol gosnmp.SnmpV3AuthProtocol
	lowerAuthProtocol := strings.ToLower(authProtocolStr)
	switch lowerAuthProtocol {
	case "":
		authProtocol = gosnmp.NoAuth
	case "md5":
		authProtocol = gosnmp.MD5
	case "sha":
		authProtocol = gosnmp.SHA
	case "sha224":
		authProtocol = gosnmp.SHA224
	case "sha256":
		authProtocol = gosnmp.SHA256
	case "sha384":
		authProtocol = gosnmp.SHA384
	case "sha512":
		authProtocol = gosnmp.SHA512
	default:
		return gosnmp.NoAuth, fmt.Errorf("unsupported authentication protocol: %s", authProtocolStr)
	}
	return authProtocol, nil
}

// GetPrivProtocol converts priv protocol from string to type
// Related resource: https://github.com/gosnmp/gosnmp/blob/f6fb3f74afc3fb0e5b44b3f60751b988bc960019/v3_usm.go#L458-L461
// Reeder AES192/256: Used by many vendors, including Cisco.
// Blumenthal AES192/256: Not many vendors use this algorithm.
func GetPrivProtocol(privProtocolStr string) (gosnmp.SnmpV3PrivProtocol, error) {
	var privProtocol gosnmp.SnmpV3PrivProtocol
	lowerPrivProtocol := strings.ToLower(privProtocolStr)
	switch lowerPrivProtocol {
	case "":
		privProtocol = gosnmp.NoPriv
	case "des":
		privProtocol = gosnmp.DES
	case "aes":
		privProtocol = gosnmp.AES
	case "aes192":
		privProtocol = gosnmp.AES192 // Blumenthal-AES192
	case "aes256":
		privProtocol = gosnmp.AES256 // Blumenthal-AES256
	case "aes192c":
		privProtocol = gosnmp.AES192C // Reeder-AES192
	case "aes256c":
		privProtocol = gosnmp.AES256C // Reeder-AES256
	default:
		return gosnmp.NoPriv, fmt.Errorf("unsupported privacy protocol: %s", privProtocolStr)
	}
	return privProtocol, nil
}

//------------------------------------------------------------------------------

// Replacer structure to store regex matching logs parts to replace.
type Replacer struct {
	Regex *regexp.Regexp
	Repl  []byte
}

// TODO: Test TraceLevelLogWriter replacements against real GoSNMP library output
//       (need more complex setup e.g. simulate gosnmp request/response)

var replacers = []Replacer{
	{
		Regex: regexp.MustCompile(`(\s*SECURITY PARAMETERS\s*:).+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*Parsed (privacyParameters|contextEngineID))\s*.+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*(AuthenticationPassphrase|PrivacyPassphrase|SecretKey|PrivacyKey|authenticationParameters)\s*:).+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*(authenticationParameters))\s*.+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*(?:Community|ContextEngineID):).+?(\s[\w]+:)`),
		Repl:  []byte(`${1}********${2}`),
	},
}

// TraceLevelLogWriter is a log writer for gosnmp logs, it removes sensitive info.
type TraceLevelLogWriter struct{}

func (sw *TraceLevelLogWriter) Write(logInput []byte) (n int, err error) {
	for _, replacer := range replacers {
		if replacer.Regex.Match(logInput) {
			logInput = replacer.Regex.ReplaceAll(logInput, replacer.Repl)
		}
	}
	l.Debugf(string(logInput))
	return len(logInput), nil
}

//------------------------------------------------------------------------------

type debugVariable struct {
	Oid      string      `json:"oid"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
	ParseErr string      `json:"parse_err,omitempty"`
}

var strippableSpecialChars = map[byte]bool{'\r': true, '\n': true, '\t': true}

// IsStringPrintable returns true if the provided byte array is only composed of printable characeters.
func IsStringPrintable(bytesValue []byte) bool {
	for _, bit := range bytesValue {
		if bit < 32 || bit > 126 {
			// The char is not a printable ASCII char but it might be a character that
			// can be stripped like `\n`
			if _, ok := strippableSpecialChars[bit]; !ok {
				return false
			}
		}
	}
	return true
}

// GetValueFromPDU converts the value from an  SnmpPDU to a standard type.
//
//nolint:lll
func GetValueFromPDU(pduVariable gosnmp.SnmpPDU) (interface{}, error) {
	switch pduVariable.Type { //nolint:exhaustive
	case gosnmp.OctetString, gosnmp.BitString:
		bytesValue, ok := pduVariable.Value.([]byte)
		if !ok {
			return nil, fmt.Errorf("oid %s: OctetString/BitString should be []byte type but got type `%T` and value `%v`", pduVariable.Name, pduVariable.Value, pduVariable.Value)
		}
		return bytesValue, nil
	case gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32, gosnmp.TimeTicks, gosnmp.Counter64, gosnmp.Uinteger32:
		return float64(gosnmp.ToBigInt(pduVariable.Value).Int64()), nil
	case gosnmp.OpaqueFloat:
		floatValue, ok := pduVariable.Value.(float32)
		if !ok {
			return nil, fmt.Errorf("oid %s: OpaqueFloat should be float32 type but got type `%T` and value `%v`", pduVariable.Name, pduVariable.Value, pduVariable.Value)
		}
		return float64(floatValue), nil
	case gosnmp.OpaqueDouble:
		floatValue, ok := pduVariable.Value.(float64)
		if !ok {
			return nil, fmt.Errorf("oid %s: OpaqueDouble should be float64 type but got type `%T` and value `%v`", pduVariable.Name, pduVariable.Value, pduVariable.Value)
		}
		return floatValue, nil
	case gosnmp.IPAddress:
		strValue, ok := pduVariable.Value.(string)
		if !ok {
			return nil, fmt.Errorf("oid %s: IPAddress should be string type but got type `%T` and value `%v`", pduVariable.Name, pduVariable.Value, pduVariable.Value)
		}
		return strValue, nil
	case gosnmp.ObjectIdentifier:
		strValue, ok := pduVariable.Value.(string)
		if !ok {
			return nil, fmt.Errorf("oid %s: ObjectIdentifier should be string type but got type `%T` and value `%v`", pduVariable.Name, pduVariable.Value, pduVariable.Value)
		}
		return strings.TrimLeft(strValue, "."), nil
	default:
		return nil, fmt.Errorf("oid %s: invalid type: %s", pduVariable.Name, pduVariable.Type.String())
	}
}

// StandardTypeToString can be used to convert the output of `GetValueFromPDU` to a string.
func StandardTypeToString(value interface{}) (string, error) {
	switch val := value.(type) {
	case float64:
		return strconv.Itoa(int(val)), nil
	case string:
		return val, nil
	case []byte:
		var strValue string
		if !IsStringPrintable(val) {
			// We hexify like Python/pysnmp impl (keep compatibility) if the value contains non ascii letters:
			// https://github.com/etingof/pyasn1/blob/db8f1a7930c6b5826357646746337dafc983f953/pyasn1/type/univ.py#L950-L953
			// hexifying like pysnmp prettyPrint might lead to unpredictable results since `[]byte` might or might not have
			// elements outside of 32-126 range. New lines, tabs and carriage returns are also stripped from the string.
			// An alternative solution is to explicitly force the conversion to specific type using profile config.
			strValue = fmt.Sprintf("%#x", val)
		} else {
			strValue = string(val)
		}
		return strValue, nil
	}
	return "", fmt.Errorf("invalid type %T for value %#v", value, value)
}

// PacketAsString used to format gosnmp.SnmpPacket for debug/trace logging.
func PacketAsString(packet *gosnmp.SnmpPacket) string {
	if packet == nil {
		return ""
	}
	var debugVariables []debugVariable
	for _, pduVariable := range packet.Variables {
		var parseError string
		name := pduVariable.Name
		value := fmt.Sprintf("%v", pduVariable.Value)
		resValue, err := GetValueFromPDU(pduVariable)
		if err == nil {
			resValueStr, err := StandardTypeToString(resValue)
			if err == nil {
				value = resValueStr
			}
		}
		if err != nil {
			parseError = fmt.Sprintf("`%s`", err)
		}
		debugVar := debugVariable{Oid: name, Type: fmt.Sprintf("%v", pduVariable.Type), Value: value, ParseErr: parseError}
		debugVariables = append(debugVariables, debugVar)
	}

	jsonPayload, err := json.Marshal(debugVariables)
	if err != nil {
		l.Debugf("error marshaling debugVar: %v", err)
		jsonPayload = []byte(``)
	}
	return fmt.Sprintf("error=%s(code:%d, idx:%d), values=%s", packet.Error, packet.Error, packet.ErrorIndex, jsonPayload)
}

//------------------------------------------------------------------------------
