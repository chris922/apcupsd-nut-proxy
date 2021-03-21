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
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
	"time"
)

func NewApcValues() *ApcValues {
	return &ApcValues{
		values:      make(map[string]string),
		refreshTime: time.Unix(0, 0),
	}
}

type IApcValues interface {
	reload(config *Config, exec execCmd) error

	get(name string) string
	getOk(name string) (string, bool)
}

type ApcValues struct {
	values map[string]string

	refreshTime time.Time
}

type execCmd func(string, ...string) ([]byte, error)

func execCommand(name string, arg ...string) ([]byte, error) {
	var out bytes.Buffer
	writer := bufio.NewWriter(&out)

	cmd := exec.Command(name, arg...)
	cmd.Stdout = writer

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "Error invoking %s", name)
	}

	return out.Bytes(), nil
}

func (ar *ApcValues) reload(config *Config, exec execCmd) error {
	out, err := exec(config.apcAccessExecutable, "-h", config.targetAddress, "-u")
	if err != nil {
		return errors.Wrapf(err, "Error invoking apcaccess")
	}

	ar.values = make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return errors.Wrapf(err, "Error reading apcaccess output")
		}

		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			// skip empty lines
			continue
		}

		pos := strings.Index(line, ":")
		if pos == -1 {
			return errors.New("Invalid line in apcaccess output")
		}

		key := strings.TrimSpace(line[:pos])
		value := strings.TrimSpace(line[(pos + 1):])

		ar.values[key] = value
	}

	ar.refreshTime = time.Now()

	return nil
}

func (av *ApcValues) get(name string) string {
	return av.values[name]
}

func (av *ApcValues) getOk(name string) (string, bool) {
	val, found := av.values[name]

	return val, found
}
