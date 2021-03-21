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
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockApcValues struct {
	mock.Mock
}

func (m *mockApcValues) reload(config *Config) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *mockApcValues) get(name string) string {
	args := m.Called(name)
	return args.String(0)
}

func (m *mockApcValues) getOk(name string) (string, bool) {
	args := m.Called(name)
	return args.String(0), args.Bool(1)
}

type responseInfo struct {
	response        string
	closeConnection bool
	errorMessage    string
}

func TestCommandReceived(t *testing.T) {
	okNoError := responseInfo{response: "OK"}

	commandToResponse := map[string]responseInfo{
		"LOGIN test":         okNoError,
		"USERNAME user":      okNoError,
		"PASSWORD password":  okNoError,
		"LOGOUT":             {response: "OK Goodbye", closeConnection: true},
		"STARTTLS":           {response: "ERR FEATURE-NOT-CONFIGURED"},
		"LIST UPS":           {response: "BEGIN LIST UPS\nUPS test \"testcase\"\nEND LIST UPS\n"},
		"LIST VAR test":      {response: "BEGIN LIST VAR test\nVAR test foo \"bar\"\nEND LIST VAR test\n"},
		"GET VAR test foo":   {response: "VAR test foo \"bar\"\n"},
		"SET VAR test model": {response: "ERR READONLY"},
	}

	apcValuesMock := &mockApcValues{}
	apcValuesMock.On("reload", mock.Anything, mock.Anything).Return(nil)
	apcValuesMock.On("getOk", "MODEL").Return("foo", true)

	for command, expResponse := range commandToResponse {
		t.Run("command="+command, func(t *testing.T) {
			response, closeConnection, err := commandReceived(command, &Config{
				upsName:        "test",
				upsDescription: "testcase",
				vars: map[string]VarLoader{
					"foo": FixedValue("bar"),
				},
			}, apcValuesMock)

			if expResponse.errorMessage == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, expResponse.errorMessage)
			}
			assert.Equal(t, expResponse.response, response)
			assert.Equal(t, expResponse.closeConnection, closeConnection)
		})
	}
}
