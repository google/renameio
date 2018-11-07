// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package renameio_test

import (
	"fmt"
	"log"

	"github.com/google/renameio"
)

func ExampleTempFile_justone() {
	persist := func(temperature float64) error {
		t, err := renameio.TempFile("", "/srv/www/metrics.txt")
		if err != nil {
			return err
		}
		defer t.Cleanup()
		if _, err := fmt.Fprintf(t, "temperature_degc %f\n", temperature); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	// Thanks to the write package, a webserver exposing /srv/www never
	// serves an incomplete or missing file.
	if err := persist(31.2); err != nil {
		log.Fatal(err)
	}
}

func ExampleTempFile_many() {
	// Prepare for writing files to /srv/www, effectively caching calls to
	// TempDir which TempFile would otherwise need to make.
	dir := renameio.TempDir("/srv/www")
	persist := func(temperature float64) error {
		t, err := renameio.TempFile(dir, "/srv/www/metrics.txt")
		if err != nil {
			return err
		}
		defer t.Cleanup()
		if _, err := fmt.Fprintf(t, "temperature_degc %f\n", temperature); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}

	// Imagine this was an endless loop, reading temperature sensor values.
	// Thanks to the write package, a webserver exposing /srv/www never
	// serves an incomplete or missing file.
	for {
		if err := persist(31.2); err != nil {
			log.Fatal(err)
		}
	}
}
