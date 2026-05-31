package filelock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type lockStub struct {
	mkdirErr    error
	openErr     error
	flockErr    error
	writePIDErr error
}

func stubLockHooks(t *testing.T, s lockStub) {
	t.Helper()
	prevMkdir := lockMkdirAll
	prevOpen := lockOpenFile
	prevFlock := lockFlock
	prevWritePID := lockWritePID
	lockMkdirAll = func(path string, perm os.FileMode) error {
		if s.mkdirErr != nil {
			return s.mkdirErr
		}
		return prevMkdir(path, perm)
	}
	lockOpenFile = func(path string, flag int, perm os.FileMode) (*os.File, error) {
		if s.openErr != nil {
			return nil, s.openErr
		}
		return prevOpen(path, flag, perm)
	}
	lockFlock = func(fd, op int) error {
		if s.flockErr != nil {
			return s.flockErr
		}
		return prevFlock(fd, op)
	}
	lockWritePID = func(f *os.File) error {
		if s.writePIDErr != nil {
			return s.writePIDErr
		}
		return prevWritePID(f)
	}
	t.Cleanup(func() {
		lockMkdirAll = prevMkdir
		lockOpenFile = prevOpen
		lockFlock = prevFlock
		lockWritePID = prevWritePID
	})
}

func TestAcquireInjectedErrors(t *testing.T) {
	errMkdir := errors.New("mkdir failed")
	errOpen := errors.New("open failed")
	errFlock := errors.New("flock failed")
	errPID := errors.New("write pid failed")

	cases := []struct {
		name string
		stub lockStub
		opts []Option
		want error
	}{
		{name: "mkdir", stub: lockStub{mkdirErr: errMkdir}, want: errMkdir},
		{name: "open", stub: lockStub{openErr: errOpen}, want: errOpen},
		{name: "flock busy", stub: lockStub{flockErr: syscall.EWOULDBLOCK}, want: ErrBusy},
		{name: "flock other", stub: lockStub{flockErr: errFlock}, want: errFlock},
		{name: "write pid", stub: lockStub{writePIDErr: errPID}, opts: []Option{WritePID()}, want: errPID},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lockPath := filepath.Join(t.TempDir(), "daemon.lock")
			stubLockHooks(t, tc.stub)
			_, err := Acquire(context.Background(), lockPath, tc.opts...)
			if tc.want == ErrBusy {
				if !errors.Is(err, ErrBusy) {
					t.Fatalf("err=%v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
		})
	}
}

func TestWithInjectedErrors(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "state.lock")
	errMkdir := errors.New("mkdir failed")

	stubLockHooks(t, lockStub{mkdirErr: errMkdir})
	if err := With(context.Background(), lockPath, func() error { return nil }); !errors.Is(err, errMkdir) {
		t.Fatalf("err=%v", err)
	}
}

func TestWritePIDOnDirectoryFails(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lockdir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := writePID(f); err == nil {
		t.Fatal("expected writePID error on directory")
	}
}

func TestWithSidecarEmptySuffixUsesDefault(t *testing.T) {
	base := filepath.Join(t.TempDir(), "state")
	if err := WithSidecar(context.Background(), base, "", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestAcquireReleaseUnlocks(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	release, err := Acquire(context.Background(), lockPath)
	if err != nil {
		t.Fatal(err)
	}
	release()
	if _, err := Acquire(context.Background(), lockPath); err != nil {
		t.Fatalf("reacquire after release: %v", err)
	}
}

func TestWritePIDHookErrors(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	f, err := os.OpenFile(lockPath, O_RDWR_CREATE, DefaultFileMode)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	errSeek := errors.New("seek failed")
	prevSeek := lockSeek
	lockSeek = func(*os.File, int64, int) (int64, error) { return 0, errSeek }
	if !errors.Is(writePID(f), errSeek) {
		t.Fatal("expected seek error")
	}
	lockSeek = prevSeek

	errTrunc := errors.New("truncate failed")
	prevTrunc := lockTruncate
	lockTruncate = func(*os.File, int64) error { return errTrunc }
	if !errors.Is(writePID(f), errTrunc) {
		t.Fatal("expected truncate error")
	}
	lockTruncate = prevTrunc

	errWrite := errors.New("write failed")
	prevWrite := lockWritePIDBytes
	lockWritePIDBytes = func(*os.File, int) error { return errWrite }
	if !errors.Is(writePID(f), errWrite) {
		t.Fatal("expected write error")
	}
	lockWritePIDBytes = prevWrite
}

func TestAcquireBlockingCancelWithStub(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	stubLockHooks(t, lockStub{flockErr: syscall.EWOULDBLOCK})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	_, err := Acquire(ctx, lockPath, Blocking())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestAcquireBlockingNonBusyError(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "daemon.lock")
	errFlock := errors.New("flock failed")
	stubLockHooks(t, lockStub{flockErr: errFlock})
	_, err := Acquire(context.Background(), lockPath, Blocking())
	if !errors.Is(err, errFlock) {
		t.Fatalf("err=%v", err)
	}
}
