package filelock

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

//
// ────────────────────────────────────────
// public lock api.
//

// Acquire takes an exclusive lock at lockPath until release is called.
// By default acquisition is non-blocking; pass [Blocking] to wait, or [WritePID]
// to record the holder process id in the lock file. Context cancellation stops a blocking wait.
func Acquire(ctx context.Context, lockPath string, opts ...Option) (release func(), err error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := lockMkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("filelock: create lock dir: %w", err)
	}
	return acquireLock(ctx, lockPath, cfg)
}

// With runs fn while holding an exclusive lock at lockPath.
// Acquisition is blocking unless [NonBlocking] is passed. Context cancellation stops a blocking wait.
func With(ctx context.Context, lockPath string, fn func() error, opts ...Option) error {
	cfg := defaultConfig()
	cfg.blocking = true
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := lockMkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return fmt.Errorf("filelock: create lock dir: %w", err)
	}
	release, err := acquireLock(ctx, lockPath, cfg)
	if err != nil {
		return err
	}
	defer release()
	return fn()
}

// WithSidecar runs fn while holding an exclusive lock at basePath+sidecar.
// When sidecar is empty, [DefaultSidecar] is used.
func WithSidecar(ctx context.Context, basePath, sidecar string, fn func() error, opts ...Option) error {
	if sidecar == "" {
		sidecar = DefaultSidecar
	}
	return With(ctx, basePath+sidecar, fn, opts...)
}

func writePID(f *os.File) error {
	if _, err := lockSeek(f, 0, 0); err != nil {
		return fmt.Errorf("filelock: seek lock file: %w", err)
	}
	if err := lockTruncate(f, 0); err != nil {
		return fmt.Errorf("filelock: truncate lock file: %w", err)
	}
	if err := lockWritePIDBytes(f, os.Getpid()); err != nil {
		return fmt.Errorf("filelock: write pid: %w", err)
	}
	return nil
}
