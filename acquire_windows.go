//go:build windows

package filelock

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

var (
	lockFileEx   = windows.LockFileEx
	unlockFileEx = windows.UnlockFileEx
)

func acquireOnce(lockPath string, cfg config) (func(), error) {
	f, err := lockOpenFile(lockPath, O_RDWR_CREATE, cfg.fileMode)
	if err != nil {
		return nil, fmt.Errorf("filelock: open lock file: %w", err)
	}
	h := windows.Handle(f.Fd())
	var ol windows.Overlapped
	lockFlags := uint32(windows.LOCKFILE_EXCLUSIVE_LOCK)
	if !cfg.blocking {
		lockFlags |= windows.LOCKFILE_FAIL_IMMEDIATELY
	}
	if err := lockFileEx(h, lockFlags, 0, 1, 0, &ol); err != nil {
		_ = f.Close()
		if !cfg.blocking && errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("filelock: lock file: %w", err)
	}
	if cfg.writePID {
		if err := lockWritePID(f); err != nil {
			_ = unlockFileEx(h, 0, 1, 0, &ol)
			_ = f.Close()
			return nil, err
		}
	}
	release := func() {
		_ = unlockFileEx(h, 0, 1, 0, &ol)
		_ = f.Close()
	}
	return release, nil
}
