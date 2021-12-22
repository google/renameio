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

package renameio

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
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

func TestChownPendingFileCreation(t *testing.T) {
	currentUid := os.Getuid()
	currentGid := os.Getgid()
	anotherGid := currentGid

	if currentUser, err := user.Current(); err != nil {
		t.Errorf("user.Current() failed: %v", err)
	} else if gids, err := currentUser.GroupIds(); err != nil {
		t.Errorf("currentUser.GroupIds() failed: %v", err)
	} else {
		anotherGidStr := gids[len(gids)-1]
		if anotherGid, err = strconv.Atoi(anotherGidStr); err != nil {
			t.Errorf("strconv.Atoi(%s) failed: %v", anotherGidStr, err)
		}
	}

	for _, tc := range []struct {
		name        string
		path        string
		options     []Option
		want        string
		wantUserID  int
		wantGroupID int
	}{
		{
			name:        "unchanged uid & gid",
			path:        filepath.Join(t.TempDir(), "modified ownership"),
			options:     []Option{WithUserID(-1), WithGroupID(-1)},
			want:        "replaced:unchanged uid & gid",
			wantUserID:  currentUid,
			wantGroupID: currentGid,
		},
		{
			name:        "changed uid & gid",
			path:        filepath.Join(t.TempDir(), "modified ownership"),
			options:     []Option{WithUserID(currentUid), WithGroupID(anotherGid)},
			want:        "replaced:changed uid & gid",
			wantUserID:  currentUid,
			wantGroupID: anotherGid,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			var pf *PendingFile

			pf, err = NewPendingFile(tc.path, tc.options...)
			if err != nil {
				t.Errorf("NewPendingFile(%q, %+v) failed: %v", tc.path, tc.options, err)
			}

			if _, err := pf.WriteString("replaced:" + tc.name); err != nil {
				t.Errorf("Write() failed: %v", err)
			}

			if err := pf.CloseAtomicallyReplace(); err != nil {
				t.Errorf("CloseAtomicallyReplace() failed: %v", err)
			}

			if _, err := pf.Write(nil); !errors.Is(err, os.ErrClosed) {
				t.Errorf("Write() after CloseAtomicallyReplace didn't fail with ErrClosed: %v", err)
			}

			if err := pf.Cleanup(); err != nil {
				t.Errorf("Cleanup() failed: %v", err)
			}

			if got, err := ioutil.ReadFile(tc.path); err != nil {
				t.Errorf("ReadFile(%q) failed: %v", tc.path, err)
			} else if string(got) != tc.want {
				t.Errorf("Read unexpected content %q from %q, want %q", string(got), tc.path, tc.want)
			}

			if fi, err := os.Stat(tc.path); err != nil {
				t.Errorf("Stat(%q) failed: %v", tc.path, err)
			} else {
				if stat, ok := fi.Sys().(*syscall.Stat_t); !ok {
					t.Errorf("stat.Sys() failed: %v", err)
				} else {
					if got := int(stat.Uid); got != tc.wantUserID {
						t.Errorf("%q has user ID %d, want %d", tc.path, got, tc.wantUserID)
					}
					if got := int(stat.Gid); got != tc.wantGroupID {
						t.Errorf("%q has group ID %d, want %d", tc.path, got, tc.wantGroupID)
					}
				}
			}
		})
	}
}
