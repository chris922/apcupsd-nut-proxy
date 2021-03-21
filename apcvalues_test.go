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
	"github.com/stretchr/testify/assert"
	"testing"
)

func testExecCommand(response string) execCmd {
	return func(name string, args ...string) ([]byte, error) {
		return []byte(response), nil
	}
}

func TestNewApcValues(t *testing.T) {
	apcValues := NewApcValues()

	assert.NotNil(t, apcValues)
	assert.NotNil(t, apcValues.values)
	assert.Equal(t, int64(0), apcValues.refreshTime.Unix())
}

func TestApcValue_reload(t *testing.T) {
	apcValues := NewApcValues()
	config := Config{}

	output := `
 STATUS : ONLINE
 UPSNAME : name
`

	apcValues.exec = testExecCommand(output)
	err := apcValues.reload(&config)
	assert.NoError(t, err)

	assert.Len(t, apcValues.values, 2)
	if assert.Contains(t, apcValues.values, "STATUS") {
		assert.Equal(t, "ONLINE", apcValues.values["STATUS"])
	}
	if assert.Contains(t, apcValues.values, "UPSNAME") {
		assert.Equal(t, "name", apcValues.values["UPSNAME"])
	}
}

func TestApcValue_get(t *testing.T) {
	apcValues := ApcValues{
		values: map[string]string{
			"key": "value",
		},
	}

	result := apcValues.get("key")

	assert.Equal(t, "value", result)
}

func TestApcValue_getOk(t *testing.T) {
	apcValues := ApcValues{
		values: map[string]string{
			"key": "value",
		},
	}

	result, found := apcValues.getOk("key")

	assert.Equal(t, "value", result)
	assert.True(t, found)

	result, found = apcValues.getOk("unknown-key")

	assert.Equal(t, "", result)
	assert.False(t, found)
}
