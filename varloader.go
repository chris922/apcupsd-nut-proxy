// Copyright [2021] [Christian Bandowski]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

// A VarLoader is a function that will be attached to NUT variables and load these values. It can access the
// configuration and apc values to retrieve the value.
type VarLoader func(name string, config *Config, av IApcValues) (string, error)

// FixedValue is a function that creates a VarLoader which always returns the given string value.
func FixedValue(value string) func(name string, config *Config, av IApcValues) (string, error) {
	return func(name string, config *Config, av IApcValues) (string, error) {
		return value, nil
	}
}

// IgnoreValue is a preconfigured FixedValue VarLoader always returning an empty string.
var IgnoreValue = FixedValue("")

// FormattedValue is a function that creates a VarLoader which accepts a format and other VarLoader of which the results
// will be used for the given format.
func FormattedValue(format string, varLoaders ...VarLoader) func(name string, config *Config,
	av IApcValues) (string, error) {

	return func(name string, config *Config, av IApcValues) (string, error) {
		values := make([]interface{}, len(varLoaders))

		for i, varLoader := range varLoaders {
			value, err := varLoader(name, config, av)
			if err != nil {
				return "", errors.WithStack(err)
			}
			values[i] = value
		}

		return fmt.Sprintf(format, values...), nil
	}
}

// ApcValue is a function that creates a VarLoader which retrieves the value of the apc values by the given apc key.
func ApcValue(apcKey string, fallback VarLoader) func(name string, config *Config, av IApcValues) (string, error) {
	return func(name string, config *Config, av IApcValues) (string, error) {
		value, ok := av.getOk(apcKey)
		if !ok {
			return fallback(name, config, av)
		}

		return value, nil
	}
}

// ApcValueMinInSec is a function that creates a VarLoader that retrieves an apc value by its key, converts it to a
// float and returns this one multiplied by 60. Assuming the apc value is in minutes, this will ensure the result is in
// minutes.
func ApcValueMinInSec(apcKey string, fallback VarLoader) func(name string, config *Config, av IApcValues) (string, error) {
	return func(name string, config *Config, av IApcValues) (string, error) {
		apcValue, err := ApcValue(apcKey, fallback)(name, config, av)
		if err != nil {
			return "", errors.WithStack(err)
		}
		if apcValue == "" {
			return "", nil
		}

		val, err := strconv.ParseFloat(apcValue, 32)
		if err != nil {
			return "", errors.Wrapf(err, "Couldn't format %s value %s as float", apcKey, apcValue)
		}

		// from minutes to seconds by multiplying with 60
		return strconv.Itoa(int(val * 60)), nil
	}
}

// the following VarLoader are there for any kind of variables that are not the same as the one e.g. available in the
// apc values, but need some extra conversion to return the response expected by NUT.

// UpsName is a VarLoader that returns the UPS name.
func UpsName(name string, config *Config, av IApcValues) (string, error) {
	return config.upsName, nil
}

// UpsDescription is a VarLoader that returns the UPS description.
func UpsDescription(name string, config *Config, av IApcValues) (string, error) {
	return config.upsDescription, nil
}

// UpsModel is a VarLoader that returns the UPS model based on the corresponding apc values.
func UpsModel(name string, config *Config, av IApcValues) (string, error) {
	value, err := ApcValue("MODEL", IgnoreValue)(name, config, av)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if value == "" {
		return "", nil
	}

	nomPowerValue, err := ApcValue("NOMPOWER", IgnoreValue)(name, config, av)
	if nomPowerValue != "" && err == nil {
		return fmt.Sprintf("%s (%s W)", value, nomPowerValue), nil
	}

	return value, nil
}

// UpsStatus is a VarLoader that returns the UPS status based on the corresponding apc values.
func UpsStatus(name string, config *Config, av IApcValues) (string, error) {
	value, err := ApcValue("STATUS", IgnoreValue)(name, config, av)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if value == "" {
		return "", nil
	}

	if strings.Contains(value, "ONLINE") {
		// use CHRG prefix in case the battery is charging (BCHARGE < 100)
		chargingValue, err := ApcValue("BCHARGE", IgnoreValue)(name, config, av)
		if chargingValue != "" && err == nil {
			chargingValueInt, err := strconv.ParseFloat(chargingValue, 32)
			if err == nil && chargingValueInt < 100.0 {
				return fmt.Sprintf("CHRG %s", value), nil
			}
		}

		return fmt.Sprintf("OL %s", value), nil
	}

	statusToResultMappings := map[string]string {
		"ONLINE": "OL",
		"ONBATT": "OB DISCHRG",
		"LOWBATT": "LB",
		"CAL": "CAL",
		"OVERLOAD": "OVER",
		"TRIM": "TRIM",
		"BOOST": "BOOST",
		"REPLACEBATT": "RB",
		"SHUTTING DOWN": "SD",
		"COMMLOST": "OFF",
	}

	result := " " + value
	for status, resultPrefix := range statusToResultMappings {
		if strings.Contains(value, status) {
			return resultPrefix + result, nil
		}
	}

	return IgnoreValue(name, config, av)
}

// UpsSelfTest is a VarLoader that returns the UPS self test results based on the corresponding apc values.
func UpsSelfTest(name string, config *Config, av IApcValues) (string, error) {
	value, err := ApcValue("SELFTEST", IgnoreValue)(name, config, av)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if value == "" {
		return "", nil
	}

	if strings.Contains(value, "OK") {
		return "OK - Battery GOOD", nil
	}
	if strings.Contains(value, "BT") {
		return "FAILED - Battery Capacity LOW", nil
	}
	if strings.Contains(value, "NG") {
		return "FAILED - Overload", nil
	}
	if strings.Contains(value, "NO") {
		return "No Test in the last 5mins", nil
	}

	return IgnoreValue(name, config, av)
}
