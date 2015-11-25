// +build linux

package ls

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func GetTermSize() (int, int, error) {
	var dimensions [4]uint16

	fd := os.Stdout.Fd()
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}

func GetLongInfo(info os.FileInfo) *longInfo {
	stat := info.Sys().(*syscall.Stat_t)
	userName := fmt.Sprintf("%d", stat.Uid)
	if u, err := userLookUp(userName); err == nil {
		userName = u
	}
	group := fmt.Sprintf("%d", stat.Gid)
	if g, err := userLookUp(group); err == nil {
		group = g
	}
	return &longInfo{userName, group, int(stat.Nlink)}
}
