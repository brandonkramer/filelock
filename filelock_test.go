package filelock_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/brandonkramer/filelock"
)

func TestAcquireExclusive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	release, err := filelock.Acquire(ctx, lockPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)

	_, err = filelock.Acquire(ctx, lockPath)
	if !errors.Is(err, filelock.ErrBusy) {
		t.Fatalf("err=%v want %v", err, filelock.ErrBusy)
	}
}

func TestAcquireWritePID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	release, err := filelock.Acquire(ctx, lockPath, filelock.WritePID())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)

	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(data))
	want := strconv.Itoa(os.Getpid())
	if got != want {
		t.Fatalf("pid=%q want %q", got, want)
	}
}

func TestWithSidecar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	base := filepath.Join(t.TempDir(), "state.json")
	called := false
	err := filelock.WithSidecar(ctx, base, filelock.DefaultSidecar, func() error {
		called = true
		if _, err := os.Stat(base + filelock.DefaultSidecar); err != nil {
			return err
		}
		return nil
	})
	if err != nil || !called {
		t.Fatalf("called=%v err=%v", called, err)
	}
}

func TestWithDirectPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "exclusive.lock")
	if err := filelock.With(ctx, lockPath, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestAcquireMkdirError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	home := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(home, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(home, "daemon.lock")
	if _, err := filelock.Acquire(ctx, lockPath); err == nil {
		t.Fatal("expected acquire error")
	}
}

func TestAcquireCanceledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	if _, err := filelock.Acquire(ctx, lockPath); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
