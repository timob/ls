// +build !cgo,!windows,!plan9 android

package ls

import (
	"errors"
	"os"
)

type LongInfo struct {
	UserName, GroupName string
	HardLinks           int
}

func GetTermSize() (int, int, error) {
	return 0, 0, errors.New("not implemented")
}

func GetLongInfo(info os.FileInfo) *LongInfo {
	return &LongInfo{"unkown", "unkown", 1}
}

func IsTerminal(fd int) bool {
	return true
}