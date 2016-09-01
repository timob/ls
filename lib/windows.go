// +build windows

package ls

import (
	"errors"
	"os"
	"os/user"
)

type LongInfo struct {
	UserName, GroupName string
	HardLinks           int
	Ino					uint64	
}

func GetTermSize() (int, int, error) {
	return 0, 0, errors.New("not implemented")
}

var userName, groupName string

func init() {
	userName = os.Getenv("USERNAME")
	if userName == "" {
		if cur, err := user.Current(); err == nil {
			userName = cur.Username
		} else {
			userName = "unknown"
		}
	}

	groupName = userName
}

func GetLongInfo(info os.FileInfo) *LongInfo {
	return &LongInfo{userName, groupName, 1, 1, 1}
}

func IsTerminal(fd int) bool {
	return true
}
