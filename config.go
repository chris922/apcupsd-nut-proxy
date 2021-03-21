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
	"flag"
	"fmt"
	"time"
)

type Config struct {
	address string
	port    int

	targetAddress string

	upsName        string
	upsDescription string

	apcAccessExecutable string

	timeout time.Duration

	vars map[string]VarLoader
}

func (c *Config) loadProgramArgs() {
	flag.StringVar(&c.address, "address", "127.0.0.1",
		"Address on which the server should listen "+
			"(use \"0.0.0.0\" to listen on all connections)")
	flag.IntVar(&c.port, "port", 3493,
		"Port number on which this server should listen")

	flag.StringVar(&c.targetAddress, "target-address", "127.0.0.1",
		"Address on which apcupsd is running")

	flag.StringVar(&c.upsName, "ups-name", "ups",
		"Name of the UPS")
	flag.StringVar(&c.upsDescription, "ups-description",
		"apcupsd NUT proxy", "Short description of the UPS")

	flag.DurationVar(&c.timeout, "timeout", time.Duration(30)*time.Second,
		"Timeout in seconds waiting for a response or sending the response. "+
			"For example \"30s\". Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\".")

	flag.StringVar(&c.apcAccessExecutable, "apcaccess-executable", "apcaccess",
		"APC Access executable")

	flag.Parse()
}

func (c Config) String() string {
	return fmt.Sprintf("Config(address=%s, port=%d, targetAddress=%s, "+
		"upsName=\"%s\", upsDescription=\"%s\", apcAccessExecutable=%s, timeout=%s)",
		c.address, c.port, c.targetAddress, c.upsName, c.upsDescription, c.timeout, c.apcAccessExecutable)
}
