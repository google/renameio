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

// +build !windows

package renameio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	for _, perm := range []os.FileMode{0o755, 0o644, 0o400, 0o765} {
		t.Run(fmt.Sprintf("perm%04o", perm), func(t *testing.T) {
			for _, umask := range []os.FileMode{0o000, 0o011, 0o007, 0o027, 0o077} {
				t.Run(fmt.Sprintf("umask%04o", umask), func(t *testing.T) {
					withUmask(t, umask)

					filename := filepath.Join(t.TempDir(), "hello.sh")

					wantData := []byte("#!/bin/sh\necho \"Hello World\"\n")
					if err := WriteFile(filename, wantData, perm); err != nil {
						t.Fatal(err)
					}

					gotData, err := ioutil.ReadFile(filename)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(gotData, wantData) {
						t.Errorf("got data %v, want data %v", gotData, wantData)
					}

					fi, err := os.Stat(filename)
					if err != nil {
						t.Fatal(err)
					}
					if gotPerm := fi.Mode() & os.ModePerm; gotPerm != perm {
						t.Errorf("got permissions %04o, want %04o", gotPerm, perm)
					}
				})
			}
		})
	}
}
