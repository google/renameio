// +build !windows

package write

import (
	"os"
)

// replaceFile atomically replaces the destination file or directory with the
// source.  It is guaranteed to either replace the target file entirely, or not
// change either file.
func replaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
