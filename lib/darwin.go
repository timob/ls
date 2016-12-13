//+build darwin

package ls

import (
	"syscall"
)

func init() {
	ioctlReadTermiosMagic = syscall.TIOCGETA
}
