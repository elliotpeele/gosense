/*
 * Copyright (c) Elliot Peele <elliot@bentlogic.net>
 *
 * This program is distributed under the terms of the MIT License as found
 * in a file called LICENSE. If it is not present, the license
 * is always available at http://www.opensource.org/licenses/mit-license.php.
 *
 * This program is distributed in the hope that it will be useful, but
 * without any warrenty; without even the implied warranty of merchantability
 * or fitness for a particular purpose. See the MIT License for full details.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/elliotpeele/gosense/record"
)

// Nagios plugin reference: https://assets.nagios.com/downloads/nagioscore/docs/nagioscore/3/en/pluginapi.html

var (
	uri       = flag.String("uri", "http://localhost:8080", "GoSense URI")
	warn_low  = flag.Float64("warn-low", 100, "Warn if value is below this point")
	warn_high = flag.Float64("warn-high", 0, "Warn if value is above this point")
	crit_low  = flag.Float64("crit-low", 100, "Critical if value is below this point")
	crit_high = flag.Float64("crit-high", 0, "Critical if value is above this point")
	age       = flag.Float64("age", 10, "Data must be no older than this value (in minutes)")
	key       = flag.String("key", "", "Sensor identifier that is being monitored, can be found at /data on the GoSense server")
)

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

func fail(rc int, rec record.Record, lh string) {
	var t string
	switch rc {
	case WARNING:
		t = "WARNING"
	case CRITICAL:
		t = "CRITICAL"
	}
	log.Printf("%s: %s.%s value too %s: %f", t, rec.Name, rec.DataType, lh, rec.Value)
	os.Exit(rc)
}

func init() {
	flag.Parse()
}

func main() {
	// Get the json representation of the sensor data.
	resp, err := http.Get(fmt.Sprintf("%s/data/%s", *uri, *key))
	if err != nil {
		log.Printf("Failed to get data: %s", err)
		os.Exit(UNKNOWN)
	}

	// Parse the json doc to a record struct
	var rec record.Record
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&rec); err != nil {
		log.Printf("Failed to parse json doc: %s", err)
		os.Exit(UNKNOWN)
	}

	// Check the age to make sure the data is current enough.
	d := time.Since(rec.TimeStamp)
	if d.Minutes() > *age {
		log.Printf("Sensor data is out of date, %f:%f", d.Minutes(), d.Seconds())
		os.Exit(UNKNOWN)
	}

	switch {
	case rec.Value < *crit_low:
		fail(CRITICAL, rec, "low")
	case rec.Value > *crit_high:
		fail(CRITICAL, rec, "high")
	case rec.Value < *warn_low:
		fail(WARNING, rec, "low")
	case rec.Value > *warn_high:
		fail(WARNING, rec, "high")
	default:
		log.Printf("OK: %s.%s value within bounds: %f", rec.Name, rec.DataType, rec.Value)
		os.Exit(OK)
	}
}
