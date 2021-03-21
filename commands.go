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
	"strings"
)

func commandReceived(command string, config *Config, apcValues IApcValues) (string, bool, error) {
	if strings.HasPrefix(command, "LOGIN ") {
		upsName := command[6:]
		if upsName != config.upsName {
			return "ERR UNKNOWN-UPS", false, nil
		}

		return "OK", false, nil
	} else if strings.HasPrefix(command, "USERNAME ") {
		// accept all usernames
		return "OK", false, nil
	} else if strings.HasPrefix(command, "PASSWORD ") {
		// accept all passwords
		return "OK", false, nil
	} else if command == "LOGOUT" {
		// close the stream
		return "OK Goodbye", true, nil
	} else if command == "STARTTLS" {
		return "ERR FEATURE-NOT-CONFIGURED", false, nil
	} else if command == "LIST UPS" {
		return commandListUps(config)
	} else if strings.HasPrefix(command, "LIST VAR ") {
		return commandListVar(command, config, apcValues)
	} else if strings.HasPrefix(command, "GET VAR ") {
		return commandGetVar(command, config, apcValues)
	} else if strings.HasPrefix(command, "SET VAR ") {
		return commandSetVar(command, config)
	} else {
		return "ERR UNKNOWN-COMMAND", false, nil
	}
}

func commandListUps(config *Config) (string, bool, error) {
	var resp strings.Builder

	resp.WriteString("BEGIN LIST UPS\n")
	resp.WriteString(fmt.Sprintf("UPS %s \"%s\"\n", config.upsName, config.upsDescription))
	resp.WriteString("END LIST UPS\n")

	return resp.String(), false, nil
}

func commandListVar(command string, config *Config, apcValues IApcValues) (string, bool, error) {
	upsName := command[9:]
	if upsName != config.upsName {
		return "ERR UNKNOWN-UPS", false, nil
	}

	err := apcValues.reload(config, execCommand)
	if err != nil {
		return "", false, errors.WithStack(err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("BEGIN LIST VAR %s\n", config.upsName))

	for name, loader := range config.vars {
		value, err := loader(name, config, apcValues)
		if err != nil {
			return "", false, errors.Wrapf(err, "Couldn't load variable %s", name)
		}
		if value == "" {
			// skip empty values
			continue
		}

		sb.WriteString(fmt.Sprintf("VAR %s %s \"%s\"\n", config.upsName, name, value))
	}

	sb.WriteString(fmt.Sprintf("END LIST VAR %s\n", config.upsName))

	return sb.String(), false, nil
}

func commandGetVar(command string, config *Config, apcValues IApcValues) (string, bool, error) {
	upsAndVarName := strings.Split(command[8:], " ")

	if len(upsAndVarName) != 2 {
		return "ERR INVALID-ARGUMENT", false, nil
	}
	if upsAndVarName[0] != config.upsName {
		return "ERR UNKNOWN-UPS", false, nil
	}
	varName := upsAndVarName[1]

	err := apcValues.reload(config, execCommand)
	if err != nil {
		return "", false, errors.WithStack(err)
	}

	loader, ok := config.vars[varName]
	if !ok {
		return "ERR VAR-NOT-SUPPORTED", false, nil
	}

	value, err := loader(varName, config, apcValues)
	if err != nil {
		return "", false, errors.Wrapf(err, "Couldn't load variable %s", varName)
	}

	return fmt.Sprintf("VAR %s %s \"%s\"\n", config.upsName, varName, value), false, nil
}

func commandSetVar(command string, config *Config) (string, bool, error) {
	upsAndVarName := strings.Split(command[8:], " ")

	if len(upsAndVarName) != 2 {
		return "ERR INVALID-ARGUMENT", false, nil
	}
	if upsAndVarName[0] != config.upsName {
		return "ERR UNKNOWN-UPS", false, nil
	}

	// we don't support writing any kind of values
	return "ERR READONLY", false, nil
}
