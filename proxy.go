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
	"github.com/pkg/errors"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func startProxy() error {
	config := Config{
		vars: map[string]VarLoader{
			"device.mfr":    UpsDescription,
			"device.model":  UpsModel,
			"device.serial": ApcValue("SERIALNO", IgnoreValue),
			"device.type":   FixedValue("ups"),

			"ups.mfr":               UpsDescription,
			"ups.mfr.date":          ApcValue("MANDATE", IgnoreValue),
			"ups.id":                FixedValue("APC"),
			"ups.vendorid":          FixedValue("051d"),
			"ups.model":             UpsModel,
			"ups.status":            UpsStatus,
			"ups.load":              ApcValue("LOADPCT", IgnoreValue),
			"ups.serial":            ApcValue("SERIALNO", IgnoreValue),
			"ups.firmware":          ApcValue("FIRMWARE", IgnoreValue),
			"ups.firmware.aux":      ApcValue("FIRMWARE", IgnoreValue),
			"ups.productid":         ApcValue("APC", IgnoreValue),
			"ups.temperature":       ApcValue("ITEMP", IgnoreValue),
			"ups.realpower.nominal": ApcValue("NOMPOWER", IgnoreValue),
			"ups.test.result":       UpsSelfTest,
			"ups.delay.start":       FixedValue("0"),
			"ups.delay.shutdown":    ApcValue("DSHUTD", IgnoreValue),
			"ups.timer.reboot":      FixedValue("-1"),
			"ups.timer.start":       FixedValue("-1"),
			"ups.timer.shutdown":    FixedValue("-1"),

			"battery.runtime":         UpsBatteryRuntime,
			"battery.runtime.low":     UpsBatteryRuntimeLow,
			"battery.charge":          ApcValue("BCHARGE", IgnoreValue),
			"battery.charge.low":      ApcValue("MBATTCHG", IgnoreValue),
			"battery.charge.warning":  FixedValue("50"),
			"battery.voltage":         ApcValue("BATTV", IgnoreValue),
			"battery.voltage.nominal": ApcValue("NOMBATTV", IgnoreValue),
			"battery.date":            ApcValue("BATTDATE", IgnoreValue),
			"battery.mfr.date":        ApcValue("BATTDATE", IgnoreValue),
			"battery.temperature":     ApcValue("ITEMP", IgnoreValue),
			"battery.type":            FixedValue("PbAc"),

			"driver.name":                   FixedValue("usbhid-ups"),
			"driver.version.internal":       FormattedValue("apcupsd %s", ApcValue("VERSION", IgnoreValue)),
			"driver.version.date":           ApcValue("DRIVER", IgnoreValue),
			"driver.parameter.pollfreq":     FixedValue("60"),
			"driver.parameter.pollinterval": FixedValue("10"),

			"input.voltage":         ApcValue("LINEV", IgnoreValue),
			"input.voltage.nominal": ApcValue("NOMINV", IgnoreValue),
			"input.sensitivity":     ApcValue("SENSE", IgnoreValue),
			"input.transfer.high":   ApcValue("HITRANS", IgnoreValue),
			"input.transfer.low":    ApcValue("LOTRANS", IgnoreValue),
			"input.frequency":       ApcValue("LINEFREQ", IgnoreValue),
			"input.transfer.reason": ApcValue("LASTXFER", IgnoreValue),

			"output.voltage":         ApcValue("OUTPUTV", IgnoreValue),
			"output.voltage.nominal": ApcValue("NOMOUTV", IgnoreValue),

			"server.info":       FixedValue("TODO"),
			"ups.beeper.status": FixedValue("enabled"),
		},
	}
	config.loadProgramArgs()

	log.Printf("Loaded configuration: %s", config)

	listenAddress := config.address + ":" + strconv.Itoa(config.port)
	l, err := net.Listen("tcp4", listenAddress)
	if err != nil {
		return errors.Wrap(err, "Couldn't start proxy")
	}
	defer l.Close()

	log.Printf("Started apcupsd NUT proxy on address %s", listenAddress)

	failedInARowCount := 0
	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("Failed accepting new connection: %s", err)
			failedInARowCount++

			if failedInARowCount >= 3 {
				return errors.Wrap(err, "Failed three times in a row accepting new connections")
			}

			continue
		}
		failedInARowCount = 0

		go handleConnection(c, &config)
	}
}

func handleConnection(c net.Conn, config *Config) {
	defer c.Close()

	log.Printf("Received request from address %s", c.RemoteAddr())

	reader := bufio.NewReader(c)
	writer := bufio.NewWriter(c)

	apcValues := NewApcValues()

	for {
		if err := c.SetDeadline(time.Now().Add(config.timeout)); err != nil {
			log.Printf("Setting the timeout for client %s failed: %+v", c.RemoteAddr(), err)
			return
		}

		command, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Reading command from client %s failed", c.RemoteAddr())
			return
		}

		command = strings.TrimSpace(command)

		log.Printf("Received command: %s", command)

		response, closeConnection, err := commandReceived(command, config, apcValues)
		if err != nil {
			log.Printf("Handling command \"%s\" for client %s failed: %+v", command, c.RemoteAddr(), err)
		}
		if response != "" {
			// ensure response ends with a newline
			response = strings.TrimSpace(response) + "\n"
			if _, err = writer.WriteString(response); err != nil {
				log.Printf("Writing response for client %s failed: %+v", c.RemoteAddr(), err)
				return
			}
		}

		if err := writer.Flush(); err != nil {
			log.Printf("Flushing response to client %s failed: %+v", c.RemoteAddr(), err)
			return
		}

		if closeConnection {
			if err = c.Close(); err != nil {
				log.Printf("Closing connection of client %s failed: %+v", c.RemoteAddr(), err)
			}

			return
		}
	}
}
