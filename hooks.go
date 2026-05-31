package filelock

import (
	"fmt"
	"os"
)

//
// ────────────────────────────────────────
// injectable hooks (tests).
//

var (
	lockMkdirAll      = os.MkdirAll
	lockOpenFile      = os.OpenFile
	lockWritePID      = writePID
	lockSeek          = func(f *os.File, off int64, whence int) (int64, error) { return f.Seek(off, whence) }
	lockTruncate      = func(f *os.File, size int64) error { return f.Truncate(size) }
	lockWritePIDBytes = func(f *os.File, pid int) error {
		_, err := fmt.Fprintf(f, "%d\n", pid)
		return err
	}
)
