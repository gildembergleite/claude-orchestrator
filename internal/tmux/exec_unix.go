//go:build !windows

package tmux

import (
	"os"
	"syscall"
)

// execReplace replaces the current process with tmux attach.
func execReplace(bin string, args ...string) error {
	argv := append([]string{bin}, args...)
	return syscall.Exec(bin, argv, os.Environ())
}
