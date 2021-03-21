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
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// mocks

func EmptyVarLoader(_ string, _ *Config, _ IApcValues) (string, error) {
	return "", nil
}
func FailingVarLoader(_ string, _ *Config, _ IApcValues) (string, error) {
	return "", errors.New("FailingVarLoader")
}
func SucceedingVarLoader(_ string, _ *Config, _ IApcValues) (string, error) {
	return "SucceedingVarLoader", nil
}
func NumberVarLoader(_ string, _ *Config, _ IApcValues) (string, error) {
	return "1", nil
}

// test cases

func TestFixedValue(t *testing.T) {
	result, err := FixedValue("foo")("name", &Config{}, &ApcValues{})

	assert.NoError(t, err)
	assert.Equal(t, "foo", result)
}

func TestApcValue(t *testing.T) {
	result, err := ApcValue("key", EmptyVarLoader)("name", &Config{}, &ApcValues{
		values: map[string]string{
			"key": "foo",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "foo", result)
}

func TestApcValue_Fallback(t *testing.T) {
	result, err := ApcValue("key", EmptyVarLoader)("name", &Config{}, &ApcValues{
		values: map[string]string{},
	})

	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestApcValue_Fallback_Error(t *testing.T) {
	result, err := ApcValue("key", FailingVarLoader)("name", &Config{}, &ApcValues{})

	assert.Equal(t, "", result)
	assert.Error(t, err)
	assert.EqualError(t, err, "FailingVarLoader")
}

func TestFormattedValue(t *testing.T) {
	result, err := FormattedValue("format %s", SucceedingVarLoader)("name", &Config{}, &ApcValues{})

	assert.NoError(t, err)
	assert.Equal(t, "format SucceedingVarLoader", result)
}

func TestUpsName(t *testing.T) {
	result, err := UpsName("name", &Config{
		upsName: "ups",
	}, &ApcValues{})

	assert.NoError(t, err)
	assert.Equal(t, "ups", result)
}

func TestUpsDescription(t *testing.T) {
	result, err := UpsDescription("name", &Config{
		upsDescription: "description",
	}, &ApcValues{})

	assert.NoError(t, err)
	assert.Equal(t, "description", result)
}

func TestUpsModel(t *testing.T) {
	result, err := UpsModel("name", &Config{}, &ApcValues{
		values: map[string]string{
			"MODEL": "model",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "model", result)
}

func TestUpsModel_WithNomPower(t *testing.T) {
	result, err := UpsModel("name", &Config{}, &ApcValues{
		values: map[string]string{
			"MODEL":    "model",
			"NOMPOWER": "300",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "model (300 W)", result)
}

func TestUpsStatus(t *testing.T) {
	statusToResult := map[string]string{
		"ONLINE": "OL ONLINE",
		"ONBATT": "OB DISCHRG ONBATT",
		"LOWBATT": "LB LOWBATT",
		"CAL": "CAL CAL",
		"OVERLOAD": "OVER OVERLOAD",
		"TRIM": "TRIM TRIM",
		"BOOST": "BOOST BOOST",
		"REPLACEBATT": "RB REPLACEBATT",
		"SHUTTING DOWN": "SD SHUTTING DOWN",
		"COMMLOST": "OFF COMMLOST",
		"UNKNOWN": "",
	}

	for status, expResult := range statusToResult {
		t.Run("STATUS=" + status, func(t *testing.T) {
			result, err := UpsStatus("name", &Config{}, &ApcValues{
				values: map[string]string{
					"STATUS": status,
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, expResult, result)
		})
	}
}

func TestUpsStatus_OnlineWithBCharge(t *testing.T) {
	result, err := UpsStatus("name", &Config{}, &ApcValues{
		values: map[string]string{
			"STATUS": "ONLINE",
			"BCHARGE": "100.0",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "OL ONLINE", result)

	result, err = UpsStatus("name", &Config{}, &ApcValues{
		values: map[string]string{
			"STATUS": "ONLINE",
			"BCHARGE": "99.9",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "CHRG ONLINE", result)
}

func TestUpsSelfTest(t *testing.T) {
	statusToResult := map[string]string{
		"OK": "OK - Battery GOOD",
		"BT": "FAILED - Battery Capacity LOW",
		"NG": "FAILED - Overload",
		"NO": "No Test in the last 5mins",
	}

	for status, expResult := range statusToResult {
		t.Run("SELFTEST=" + status, func(t *testing.T) {
			result, err := UpsSelfTest("name", &Config{}, &ApcValues{
				values: map[string]string{
					"SELFTEST": status,
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, expResult, result)
		})
	}
}

func TestApcValueMinInSec(t *testing.T) {
	statusToResult := map[string]string{
		"1": "60",
		"10": "600",
		"1.5": "90",
	}

	for status, expResult := range statusToResult {
		t.Run("VALUE=" + status, func(t *testing.T) {
			result, err := ApcValueMinInSec("VALUE", EmptyVarLoader)("name", &Config{}, &ApcValues{
				values: map[string]string{
					"VALUE": status,
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, expResult, result)
		})
	}
}

func TestApcValueMinInSec_InvalidNumber(t *testing.T) {
	result, err := ApcValueMinInSec("VALUE", NumberVarLoader)("name", &Config{}, &ApcValues{
		values: map[string]string{
			"VALUE": "not-a-number",
		},
	})

	assert.Equal(t, "", result)
	assert.Error(t, err)
	assert.EqualError(t, err, "Couldn't format VALUE value not-a-number as float: " +
		"strconv.ParseFloat: parsing \"not-a-number\": invalid syntax")
}

func TestApcValueMinInSec_Fallback(t *testing.T) {
	result, err := ApcValueMinInSec("VALUE", NumberVarLoader)("name", &Config{}, &ApcValues{
		values: map[string]string{},
	})

	assert.NoError(t, err)
	assert.Equal(t, "60", result)
}
