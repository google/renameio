// Copyright 2019 Google Inc.
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

// Command tempdircache_test tests TempDirCache.
//
// For each dest (passed as command line arguments), it prints the dest, the
// result of renameio.TempDir(dest), and the result of TempDirCache.Get(dest).
// When given multiple dests are on the same filesystem but on a different
// filesystem to the result of os.TempDir(), TempDirCache.Get(dest) should
// repeatedly return the first result of renameio.TempDir(dest), whereas
// renameio.TempDir(dest) should return different values.
package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/google/renameio"
)

func main() {
	flag.Parse()
	tempDirCache := renameio.NewTempDirCache()
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(tw, "dest\trenameio.TempDir(dest)\tTempDirCache.Get(dest)\n")
	for _, dest := range flag.Args() {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", dest, renameio.TempDir(dest), tempDirCache.Get(dest))
	}
	tw.Flush()
}
