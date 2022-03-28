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
	sinks := []map[string]interface{}{}

	if len(categoryShorts) != len(args) || len(categoryShorts) == 0 {
		return nil, fmt.Errorf("programming error")
	}

	for k, v := range categoryShorts {
		if err := polymerizeSinkCategory(v, args[k], &sinks); err != nil {
			return nil, err
		}
	}

	return sinks, nil
}

const sepSinks = "||"

func polymerizeSinkCategory(categoryShort, arg string, sinks *[]map[string]interface{}) error {
	if len(arg) == 0 {
		return nil
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
				existID, err := dkstring.GetMapMD5String(existSink)
				if err != nil {
					return err
				}

				getID, err := dkstring.GetMapMD5StringX(mSingle)
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

func parseSinkSingle(single string) (map[string]string, error) {
	// single: influxdb://1.1.1.1:8086?database=db0&timeout=15s
	if single == "" {
		return nil, nil
	}

	mSingle := make(map[string]string)

	uURL, err := url.Parse(single)
	if err != nil {
		return nil, err
	}

	if uURL.Scheme == "" {
		return nil, fmt.Errorf("invalid scheme")
	}

	mSingle["target"] = uURL.Scheme
	mSingle["host"] = uURL.Host

	mSub, err := url.ParseQuery(uURL.RawQuery)
	if err != nil {
		return nil, err
	}

	for k, v := range mSub {
		if len(v) == 0 {
			continue
		}
		mSingle[k] = v[0]
	}

	return mSingle, nil
}
