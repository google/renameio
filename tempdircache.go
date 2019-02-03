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

package renameio

// A TempDirCache caches calls to TempDir. Destinations on the same device
// share the same temporary directory.
type TempDirCache struct {
	cache map[int]string
}

// NewTempDirCache returns a new TempDirCache.
func NewTempDirCache() *TempDirCache {
	return &TempDirCache{
		cache: make(map[int]string),
	}
}

// Get is the equivalent of TempDir, except using c is a cache.
func (c *TempDirCache) Get(dest string) string {
	dev, devOK := getDev(dest)
	if devOK {
		tempDir, ok := c.cache[dev]
		if ok {
			return tempDir
		}
	}
	tempDir := TempDir(dest)
	if devOK {
		c.cache[dev] = tempDir
	}
	return tempDir
}
