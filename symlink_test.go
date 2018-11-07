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

func TestSymlink(t *testing.T) {
	d, err := ioutil.TempDir("", "tempdirtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	want := []byte("Hello World")
	if err := ioutil.WriteFile(filepath.Join(d, "hello.txt"), want, 0644); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := Symlink("hello.txt", filepath.Join(d, "hi.txt")); err != nil {
			t.Fatal(err)
		}

		got, err := ioutil.ReadFile(filepath.Join(d, "hi.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("unexpected content: got %q, want %q", string(got), string(want))
		}
	}
}
