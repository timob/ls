// +build darwin dragonfly freebsd !android,linux netbsd openbsd solaris
// +build cgo

package ls

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"
)

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <grp.h>
#include <stdlib.h>
#include <string.h>
#include <locale.h>

static int mygetgrgid_r(int gid, struct group *grp,
	char *buf, size_t buflen, struct group **result) {
	return getgrgid_r(gid, grp, buf, buflen, result);
}
*/
import "C"

type UnknownGroupIdError int

func (e UnknownGroupIdError) Error() string {
	return "user: unknown group id " + strconv.Itoa(int(e))
}

// taken from os/user/lookup_unix.go
func groupNameOSLookup(gid int) (string, error) {
	var grp C.struct_group
	var result *C.struct_group

	var bufSize C.long
	if runtime.GOOS == "dragonfly" || runtime.GOOS == "freebsd" {
		// DragonFly and FreeBSD do not have _SC_GETPW_R_SIZE_MAX
		// and just return -1.  So just use the same
		// size that Linux returns.
		bufSize = 1024
	} else {
		bufSize = C.sysconf(C._SC_GETGR_R_SIZE_MAX)
		if bufSize <= 0 || bufSize > 1<<20 {
			return "", fmt.Errorf("user: unreasonable _SC_GETGR_R_SIZE_MAX of %d", bufSize)
		}
	}
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)
	var rv C.int
	// mygetgrgid_r is a wrapper around getgrgid_r to
	// to avoid using gid_t because C.gid_t(uid) for
	// unknown reasons doesn't work on linux.
	rv = C.mygetgrgid_r(C.int(gid),
		&grp,
		(*C.char)(buf),
		C.size_t(bufSize),
		&result)
	if rv != 0 {
		return "", fmt.Errorf("ls: lookup group failed id %d: %s", gid, syscall.Errno(rv))
	}
	if result == nil {
		return "", UnknownGroupIdError(gid)
	}
	return C.GoString(grp.gr_name), nil
}

var groupLookupCache = make(map[string]string)

func groupLookup(id string) (string, error) {
	if v, ok := groupLookupCache[id]; ok {
		return v, nil
	} else {
		i, e := strconv.Atoi(id)
		if e != nil {
			return "", e
		}
		g, err := groupNameOSLookup(i)
		if err == nil {
			groupLookupCache[id] = g
			return g, nil
		}
		return "", err
	}
}

var userLookupCache = make(map[string]string)

func userLookUp(id string) (string, error) {
	if v, ok := userLookupCache[id]; ok {
		return v, nil
	} else {
		u, err := user.LookupId(id)
		if err == nil {
			userLookupCache[id] = u.Username
			return u.Username, nil
		}
		return "", err
	}
}

type LongInfo struct {
	UserName, GroupName string
	HardLinks           int
	Ino					uint64	
}

func GetTermSize() (int, int, error) {
	var dimensions [4]uint16

	fd := os.Stdin.Fd()
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}

var ioctlReadTermiosMagic uintptr

func init() {
	if runtime.GOOS != "darwin" {
		ioctlReadTermiosMagic = 0x5401
	}
}

func IsTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermiosMagic , uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

var setLocaleCalled bool

func Strcoll(s1, s2 string) int {
	if setLocaleCalled == false {
		cstr := C.CString("")
		C.setlocale(C.LC_ALL, cstr)
		C.free(unsafe.Pointer(cstr))
		setLocaleCalled = true
	}
	cs1 := C.CString(s1)
	cs2 := C.CString(s2)
	defer C.free(unsafe.Pointer(cs1))
	defer C.free(unsafe.Pointer(cs2))
	return int(C.strcoll(cs1, cs2))
}

func GetLongInfo(info os.FileInfo) *LongInfo {
	stat := info.Sys().(*syscall.Stat_t)
	userName := fmt.Sprintf("%d", stat.Uid)
	if u, err := userLookUp(userName); err == nil {
		userName = u
	}
	group := fmt.Sprintf("%d", stat.Gid)
	if g, err := groupLookup(group); err == nil {
		group = g
	}
	return &LongInfo{userName, group, int(stat.Nlink), stat.Ino}
}
