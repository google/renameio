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

//go:build !windows
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

					maskedPerm := perm & ^umask

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
					if gotPerm := fi.Mode() & os.ModePerm; gotPerm != maskedPerm {
						t.Errorf("got permissions %04o, want %04o", gotPerm, maskedPerm)
					}
				})
			}
		})
	}
}

func TestWriteFileIgnoreUmask(t *testing.T) {
	withUmask(t, 0o077)

	filename := filepath.Join(t.TempDir(), "file")

	const wantPerm os.FileMode = 0o765

	if err := WriteFile(filename, nil, wantPerm, IgnoreUmask()); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if gotPerm := fi.Mode() & os.ModePerm; gotPerm != wantPerm {
		t.Errorf("got permissions %04o, want %04o", gotPerm, wantPerm)
	}
}

func TestWriteFileEquivalence(t *testing.T) {
	type writeFunc func(string, []byte, os.FileMode, ...Option) error
	type test struct {
		name   string
		fn     writeFunc
		perm   os.FileMode
		umask  os.FileMode
		exists bool
	}

	var tests []test

	for _, wf := range []struct {
		name string
		fn   writeFunc
	}{
		{
			name: "WriteFile",
			fn:   WriteFile,
		},
		{
			name: "ioutil",
			fn: func(filename string, data []byte, perm os.FileMode, opts ...Option) error {
				return ioutil.WriteFile(filename, data, perm)
			},
		},
	} {
		for _, perm := range []os.FileMode{0o755, 0o644, 0o400, 0o765} {
			for _, umask := range []os.FileMode{0o000, 0o011, 0o007, 0o027, 0o077} {
				for _, exists := range []bool{false, true} {
					name := fmt.Sprintf("%s/perm%04o/umask%04o", wf.name, perm, umask)
					if exists {
						name += "/exists"
					}

					tests = append(tests, test{
						name:   name,
						fn:     wf.fn,
						perm:   perm,
						umask:  umask,
						exists: exists,
					})
				}
			}
		}
	}

	const existingPerm os.FileMode = 0o654

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withUmask(t, tc.umask)

			maskedPerm := tc.perm & ^tc.umask

			filename := filepath.Join(t.TempDir(), "test.txt")

			if tc.exists {
				// Create file in preparation for replacement
				fh, err := os.Create(filename)
				if err != nil {
					t.Errorf("Create(%q) failed: %v", filename, err)
				}

				if err := fh.Chmod(existingPerm); err != nil {
					t.Errorf("Chmod() failed: %v", err)
				}

				fh.Close()

				maskedPerm = existingPerm
			}

			wantData := []byte("content\n")

			if err := tc.fn(filename, wantData, tc.perm); err != nil {
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
			if gotPerm := fi.Mode() & os.ModePerm; gotPerm != maskedPerm {
				t.Errorf("got permissions %04o, want %04o", gotPerm, maskedPerm)
			}
		})
	}
}
