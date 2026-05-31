//go:build unix

package filelock

import (
	"errors"
	"fmt"
	"syscall"
)

var lockFlock = syscall.Flock

func acquireOnce(lockPath string, cfg config) (func(), error) {
	f, err := lockOpenFile(lockPath, O_RDWR_CREATE, cfg.fileMode)
	if err != nil {
		return nil, fmt.Errorf("filelock: open lock file: %w", err)
	}
	flags := syscall.LOCK_EX
	if !cfg.blocking {
		flags |= syscall.LOCK_NB
	}
	if err := lockFlock(int(f.Fd()), flags); err != nil {
		_ = f.Close()
		if !cfg.blocking && errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("filelock: flock: %w", err)
	}
	if cfg.writePID {
		if err := lockWritePID(f); err != nil {
			_ = lockFlock(int(f.Fd()), syscall.LOCK_UN)
			_ = f.Close()
			return nil, err
		}
	}
	release := func() {
		_ = lockFlock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
	return release, nil
}
