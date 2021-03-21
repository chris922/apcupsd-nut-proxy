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

type VarLoader func(name string, config *Config, av IApcValues) (string, error)

func FixedValue(value string) func(name string, config *Config, av IApcValues) (string, error) {
	return func(name string, config *Config, av IApcValues) (string, error) {
		return value, nil
	}
}

var IgnoreValue = FixedValue("")

func FormattedValue(format string, varLoaders ...VarLoader) func(name string, config *Config, av IApcValues) (string, error) {
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

func ApcValue(apcKey string, fallback VarLoader) func(name string, config *Config, av IApcValues) (string, error) {
	return func(name string, config *Config, av IApcValues) (string, error) {
		value, ok := av.getOk(apcKey)
		if !ok {
			return fallback(name, config, av)
		}

		return value, nil
	}
}

func UpsName(name string, config *Config, av IApcValues) (string, error) {
	return config.upsName, nil
}

func UpsDescription(name string, config *Config, av IApcValues) (string, error) {
	return config.upsDescription, nil
}

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
/*UPS_STATUS=""

if [[ $VALUE == *"ONLINE"* ]]; then
								UPS_STATUS="OL $UPS_STATUS";
								if [ $BATTNOTFULL = 1 ]; then UPS_STATUS="CHRG $UPS_STATUS"; fi
							   fi
if [[ $VALUE == *"ONBATT"* ]]; then UPS_STATUS="OB DISCHRG $UPS_STATUS"; fi
if [[ $VALUE == *"LOWBATT"* ]]; then UPS_STATUS="LB $UPS_STATUS"; fi
if [[ $VALUE == *"CAL"* ]]; then UPS_STATUS="CAL $UPS_STATUS"; fi
if [[ $VALUE == *"OVERLOAD"* ]]; then UPS_STATUS="OVER $UPS_STATUS"; fi
if [[ $VALUE == *"TRIM"* ]]; then UPS_STATUS="TRIM $UPS_STATUS"; fi
if [[ $VALUE == *"BOOST"* ]]; then UPS_STATUS="BOOST $UPS_STATUS"; fi
if [[ $VALUE == *"REPLACEBATT"* ]]; then UPS_STATUS="RB $UPS_STATUS"; fi
if [[ $VALUE == *"SHUTTING DOWN"* ]]; then UPS_STATUS="SD $UPS_STATUS"; fi
if [[ $VALUE == *"COMMLOST"* ]]; then UPS_STATUS="OFF $UPS_STATUS"; fi
UPS_STATUS="$(echo -e "${UPS_STATUS}" | sed -e 's/[[:space:]]*$//')"*/

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
/*SELFTEST)	if [[ $VALUE == *"OK"* ]]; then UPS_SELFTEST="OK - Battery GOOD";
elif [[ $VALUE == *"BT"* ]]; then UPS_SELFTEST="FAILED - Battery Capacity LOW";
elif [[ $VALUE == *"NG"* ]]; then UPS_SELFTEST="FAILED - Overload";
elif [[ $VALUE == *"NO"* ]]; then UPS_SELFTEST="No Test in the last 5mins";
fi;;*/

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

var UpsBatteryRuntime = ApcValueMinInSec("TIMELEFT", IgnoreValue)
/*TIMELEFT)	let UPS_TIMELEFT="$(echo -e "${VALUE}" | cut -d'.' -f1)"*60;;	#only use string before ".", multiply with 60 for value in seconds*/

var UpsBatteryRuntimeLow = ApcValueMinInSec("DLOWBATT", IgnoreValue)
/*DLOWBATT)	let UPS_DLOWBATT="$(echo -e "${VALUE}" | cut -d'.' -f1)"*60;;	#low battery runtime [min] * 60 for seconds*/
