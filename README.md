# filelock

Cross-platform advisory file locks for single-instance services, registries, and on-disk critical sections.

## Install

From [pkg.go.dev](https://pkg.go.dev/github.com/brandonkramer/filelock):

```bash
go get github.com/brandonkramer/filelock
```

## Quick start

```go
// Daemon single-instance lock (non-blocking + PID marker)
release, err := filelock.Acquire(context.Background(), "/var/run/myapp/daemon.lock", filelock.WritePID())
if err != nil {
    return err
}
defer release()

// Registry / state file sidecar lock
err = filelock.WithSidecar(context.Background(), "/var/lib/myapp/state.json", filelock.DefaultSidecar, func() error {
    return updateState()
})
```

## Options

| Option | Use with | Effect |
|--------|----------|--------|
| `WritePID()` | `Acquire`, `With` | Write holder PID into lock file |
| `Blocking()` | `Acquire` | Wait for lock instead of returning `ErrBusy` |
| `NonBlocking()` | `With` | Try once; return `ErrBusy` when held |
| `FileMode(mode)` | both | Permission bits for new lock files |

## Development

```bash
make check
```
