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

// +build linux

package renameio

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestTempDir(t *testing.T) {
	if tmpdir, ok := os.LookupEnv("TMPDIR"); ok {
		defer os.Setenv("TMPDIR", tmpdir) // restore
	} else {
		defer os.Unsetenv("TMPDIR") // restore
	}

	mount1, err := ioutil.TempDir("", "tempdirtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mount1)

	mount2, err := ioutil.TempDir("", "tempdirtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mount2)

	if err := syscall.Mount("tmpfs", mount1, "tmpfs", 0, ""); err != nil {
		t.Skipf("cannot mount tmpfs on %s: %v", mount1, err)
	}
	defer syscall.Unmount(mount1, 0)

	if err := syscall.Mount("tmpfs", mount2, "tmpfs", 0, ""); err != nil {
		t.Skipf("cannot mount tmpfs on %s: %v", mount2, err)
	}
	defer syscall.Unmount(mount2, 0)

	tests := []struct {
		name   string
		dir    string
		path   string
		TMPDIR string
		want   string
	}{
		{
			name: "implicit TMPDIR",
			path: filepath.Join(os.TempDir(), "foo.txt"),
			want: os.TempDir(),
		},

		{
			name:   "explicit TMPDIR",
			path:   filepath.Join(mount1, "foo.txt"),
			TMPDIR: mount1,
			want:   mount1,
		},

		{
			name:   "explicit unsuitable TMPDIR",
			path:   filepath.Join(mount1, "foo.txt"),
			TMPDIR: mount2,
			want:   mount1,
		},

		{
			name:   "nonexistant TMPDIR",
			path:   filepath.Join(mount1, "foo.txt"),
			TMPDIR: "/nonexistant",
			want:   mount1,
		},

		{
			name:   "caller-specified",
			dir:    "/overridden",
			path:   filepath.Join(mount1, "foo.txt"),
			TMPDIR: "/nonexistant",
			want:   "/overridden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.TMPDIR == "" {
				os.Unsetenv("TMPDIR")
			} else {
				os.Setenv("TMPDIR", tt.TMPDIR)
			}
			if got := tempDir(tt.dir, tt.path); got != tt.want {
				t.Fatalf("tempDir(%q, %q): got %q, want %q", tt.dir, tt.path, got, tt.want)
			}
		})
	}
}
