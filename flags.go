package filelock

import "os"

// O_RDWR_CREATE is the open flag set used for lock files.
const O_RDWR_CREATE = os.O_CREATE | os.O_RDWR
