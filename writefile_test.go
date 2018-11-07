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

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package renameio

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	d, err := ioutil.TempDir("", "tempdirtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	filename := filepath.Join(d, "hello.sh")

	wantData := []byte("#!/bin/sh\necho \"Hello World\"\n")
	wantPerm := os.FileMode(0755)
	if err := WriteFile(filename, wantData, wantPerm); err != nil {
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
	if gotPerm := fi.Mode() & os.ModePerm; gotPerm != wantPerm {
		t.Errorf("got permissions 0%o, want permissions 0%o", gotPerm, wantPerm)
	}
}
