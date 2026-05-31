package filelock_test

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/brandonkramer/filelock"
)

func TestAcquireBlockingWaitsSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		lockPath := filepath.Join(t.TempDir(), "daemon.lock")
		release, err := filelock.Acquire(ctx, lockPath)
		if err != nil {
			t.Fatal(err)
		}

		acquired := make(chan error, 1)
		go func() {
			rel, err := filelock.Acquire(ctx, lockPath, filelock.Blocking())
			if err != nil {
				acquired <- err
				return
			}
			rel()
			acquired <- nil
		}()

		synctest.Wait()
		release()
		time.Sleep(20 * time.Millisecond)
		synctest.Wait()
		if err := <-acquired; err != nil {
			t.Fatal(err)
		}
	})
}

func TestAcquireBlockingContextCanceledSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		lockPath := filepath.Join(t.TempDir(), "daemon.lock")
		release, err := filelock.Acquire(ctx, lockPath)
		if err != nil {
			t.Fatal(err)
		}
		defer release()

		waitCtx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			_, err := filelock.Acquire(waitCtx, lockPath, filelock.Blocking())
			done <- err
		}()

		synctest.Wait()
		cancel()
		time.Sleep(20 * time.Millisecond)
		synctest.Wait()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestWithNonBlockingBusy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "state.lock")
	release, err := filelock.Acquire(ctx, lockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	err = filelock.With(ctx, lockPath, func() error { return nil }, filelock.NonBlocking())
	if !errors.Is(err, filelock.ErrBusy) {
		t.Fatalf("err=%v", err)
	}
}

func TestWithReturnsFnError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "state.lock")
	fnErr := errors.New("fn failed")
	err := filelock.With(ctx, lockPath, func() error { return fnErr })
	if !errors.Is(err, fnErr) {
		t.Fatalf("err=%v", err)
	}
}

func TestAcquireFileMode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	release, err := filelock.Acquire(ctx, lockPath, filelock.FileMode(0o600))
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestConcurrentAcquireRelease(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				release, err := filelock.Acquire(ctx, lockPath)
				if errors.Is(err, filelock.ErrBusy) {
					continue
				}
				if err != nil {
					t.Errorf("acquire: %v", err)
					return
				}
				release()
			}
		})
	}
	wg.Wait()
}
