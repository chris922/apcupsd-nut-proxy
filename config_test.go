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
	"time"
)

func TestConfig_loadProgramArgs(t *testing.T) {
	config := &Config{}
	config.loadProgramArgs()

	assert.Equal(t, "127.0.0.1", config.address)
	assert.Equal(t, 3493, config.port)
	assert.Equal(t, "127.0.0.1", config.targetAddress)
	assert.Equal(t, "ups", config.upsName)
	assert.Equal(t, "apcupsd NUT proxy", config.upsDescription)
	assert.Equal(t, "apcaccess", config.apcAccessExecutable)
	assert.Equal(t, time.Duration(30) * time.Second, config.timeout)
	assert.Nil(t, config.vars)
}

func TestConfig_String(t *testing.T) {
	config := &Config{
		address:             "address",
		port:                1000,
		targetAddress:       "targetAddress",
		upsName:             "upsName",
		upsDescription:      "upsDescription",
		apcAccessExecutable: "apcAccessExecutable",
		timeout:             42,
		vars:                nil,
	}

	result := config.String()

	assert.Contains(t, result, "address")
	assert.Contains(t, result, "1000")
	assert.Contains(t, result, "targetAddress")
	assert.Contains(t, result, "upsName")
	assert.Contains(t, result, "upsDescription")
	assert.Contains(t, result, "apcAccessExecutable")
	assert.Contains(t, result, "42")
}
