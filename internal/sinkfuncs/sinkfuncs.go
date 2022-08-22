// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkfuncs contains sink's general functions
package sinkfuncs

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
)

func GetSinkFromEnvs(categoryShorts, args []string) ([]map[string]interface{}, error) {
	sinks := []map[string]interface{}{{}} // default from config.DefaultConfig

	if len(categoryShorts) != len(args) || len(categoryShorts) == 0 {
		return nil, fmt.Errorf("unmatched category length(%d <> %d)", len(categoryShorts), len(args))
	}

	for k, v := range categoryShorts {
		if err := polymerizeSinkCategory(v, args[k], &sinks); err != nil {
			return nil, err
		}
	}

	return sinks, nil
}

func GetSinkCreatorID(mVal map[string]interface{}) (string, string, error) {
	return dkstring.GetMapMD5String(mVal, excludeKeyName)
}

//------------------------------------------------------------------------------

const sepSinks = "||"

var excludeKeyName = []string{
	"categories",
}

func polymerizeSinkCategory(categoryShort, arg string, sinks *[]map[string]interface{}) error {
	if len(arg) == 0 {
		return nil
	}

	// reset if needed
	if len(*sinks) == 0 || len((*sinks)[0]) == 0 {
		(*sinks) = []map[string]interface{}{} // caution: not default []map[string]interface{}{{}}
	}

	// arg: influxdb://1.1.1.1:8086?database=db0&timeout=15s||influxdb://1.1.1.1:8087?database=db0&timeout=15s
	sks := strings.Split(arg, sepSinks)
	for _, sk := range sks {
		// sk: influxdb://1.1.1.1:8086?database=db0&timeout=15s
		mSingle, err := parseSinkSingle(sk)
		if err != nil {
			return err
		}

		if len(mSingle) == 0 {
			continue
		}

		foundID := false
		for k, existSink := range *sinks {
			targetInterface, ok := existSink["target"]
			if !ok {
				return fmt.Errorf("not found target")
			}
			targetString, ok := targetInterface.(string)
			if !ok {
				return fmt.Errorf("target not string")
			}
			if targetString == mSingle["target"] {
				existID, _, err := GetSinkCreatorID(existSink)
				if err != nil {
					return err
				}
				getID, _, err := GetSinkCreatorID(mSingle)
				if err != nil {
					return err
				}

				if getID == existID {
					foundID = true

					// append category
					categoriesInterface, ok := existSink["categories"]
					if !ok {
						return fmt.Errorf("not found categories")
					}

					categoriesArray, ok := categoriesInterface.([]string)
					if !ok {
						return fmt.Errorf("categories not []string: %s, %s", reflect.TypeOf(categoriesInterface).Name(), reflect.TypeOf(categoriesInterface).String()) //nolint:lll
					}

					foundCategory := false
					for _, existCategory := range categoriesArray {
						existCategoryUpper := strings.ToUpper(existCategory)
						if categoryShort == existCategoryUpper {
							// category already exist
							foundCategory = true
							break
						}
					}
					if !foundCategory {
						// append new category
						categoriesArray = append(categoriesArray, categoryShort)
						(*sinks)[k]["categories"] = categoriesArray
					}

					break
				}
			}
		}
		if !foundID {
			// append all
			newMapInterface := make(map[string]interface{})
			for k, v := range mSingle {
				newMapInterface[k] = v
				newMapInterface["categories"] = []string{categoryShort}
			}
			(*sinks) = append((*sinks), newMapInterface)
		}
	}

	return nil
}

func parseSinkSingle(single string) (map[string]interface{}, error) {
	// single: influxdb://1.1.1.1:8086?database=db0&timeout=15s
	if single == "" {
		return nil, nil
	}

	mSingle := make(map[string]interface{})

	uURL, err := url.Parse(single)
	if err != nil {
		return nil, err
	}

	if uURL.Scheme == "" {
		return nil, fmt.Errorf("invalid scheme")
	}

	mSingle["target"] = uURL.Scheme
	if len(uURL.Host) > 0 {
		mSingle["host"] = uURL.Host
	}

	mSub, err := url.ParseQuery(uURL.RawQuery)
	if err != nil {
		return nil, err
	}

	for k, v := range mSub {
		if len(v) == 0 {
			continue
		}
		if k == "filters" {
			mSingle[k] = v
		} else {
			mSingle[k] = v[0]
		}
	}

	return mSingle, nil
}
