package main

import (
	"os"
	"os/user"
	"github.com/timob/list"
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"
	"path"
	"github.com/bradfitz/slice"
	"time"
)

type DisplayEntry struct {
	path string
	os.FileInfo
}

type DisplayEntryList struct {
	Data []DisplayEntry
	list.Slice
}

var now = time.Now()

var showDirEntries bool
var showAll bool
var showAlmostAll bool
var longList bool
const (
	name int = iota
	modTime int = iota
	size int = iota
)
var sortType int = name
var reverseSort bool
var humanReadable bool
var recursiveList bool
var onlyHidden bool

var width int


func getTermSize() (int, int, error) {
	var dimensions [4]uint16

	fd := os.Stdout.Fd()
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}


func human(n int64) string {
	var i int64
	var w = n
	var uSize int64 = 1
	for i = 0; i < 12; i++ {
		if w / 1024 == 0 {
			break
		}
		w = w / 1024
		uSize *= 1024
	}

	var f int64
	var unit string
	f = (n - w * uSize)
	if f != 0 && i > 0{
		if w < 10 {
			if f != 0 {
				lowerSize := uSize / 1024
				// magic plus one here (seems to be what GNU ls does)
				tenth := int64(1024 / 10) + 1
				f = f / lowerSize / tenth
				// round up
				f++
				if f == 10 {
					w++
					f = 0
				}
			}
		} else {
			// round up
			w++
		}
	}

	switch i {
	case 1:
		unit = "K"
	case 2:
		unit = "M"
	case 3:
		unit = "G"
	case 4:
		unit = "T"
	}

	if w > 0 && w < 10 {
		return fmt.Sprintf("%d.%d%s", w, f, unit)
	} else {
		return fmt.Sprintf("%d%s", w, unit)
	}
}

func decimalLen(n int64) (i int) {
	for i = 1; i < 24; i++ {
		if n / 10 == 0 {
			break
		}
		n = n / 10
	}
	return
}

var userLookupCache = make(map[string]string)
func userLookUp(id string) (string, error) {
	if v, ok := userLookupCache[id]; ok {
		return v, nil
	} else {
		u, err := user.LookupId(id)
		if err == nil {
			userLookupCache[id] = u.Name
			return u.Name, nil
		}
		return "", err
	}
}

type longInfo struct {
	userName, groupName string
	hardLinks int
}

func getLongInfo(info os.FileInfo) *longInfo {
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


func strcmpi(a, b string) int {
	for i, av := range a {
		if i >= len(b) {
			return 1
		}
		if av > 96 && av < 123 {
			av -= 32
		}
		bv := rune(b[i])
		if bv > 96 && bv < 123 {
			bv -= 32
		}

		if av != bv {
			if av > bv {
				return 1
			} else {
				return -1
			}
		}
	}

	if len(b) > len(a) {
		return -1
	} else {
		return 0
	}
}

func modeString(mode os.FileMode) string {
	output := []byte(strings.Repeat("-", 10))
	if mode & os.ModeDir != 0{
		output[0] = 'd'
	} else if mode & os.ModeSymlink != 0{
		output[0] = 'l'
	} else if mode & os.ModeNamedPipe != 0{
		output[0] = 'p'
	} else if mode & os.ModeSocket != 0{
		output[0] = 's'
	} else if mode & os.ModeCharDevice != 0 && mode & os.ModeDevice != 0 {
		output[0] = 'c'
	}

	const rwx = "rwxrwxrwx"
	for i, c := range rwx {
		bitSet := mode&(1<<uint(9-1-i)) != 0
		if bitSet {
			if (i == 2 && mode & os.ModeSetuid != 0) || (i == 5 && mode & os.ModeSetgid != 0) {
				output[i+1] = 's'
			} else if i == 8 && mode & os.ModeSticky != 0 {
				output[i+1] = 't'
			} else {
				output[i+1] = byte(c)
			}
		} else if (i == 2 && mode & os.ModeSetuid != 0) || (i == 5 && mode & os.ModeSetgid != 0) {
			output[i+1] = 'S'
		} else if i == 8 && mode & os.ModeSticky != 0 {
			output[i+1] = 'T'
		}
	}

	return string(output)
}

func display(selected []DisplayEntry, root string) {
	slice.Sort(selected, func(i, j int) (v bool) {
		var same bool
		if sortType == modTime {
			v = selected[i].ModTime().Before(selected[j].ModTime())
			if !v {
				same = selected[i].ModTime().Equal(selected[	j].ModTime())
			}
			v = !v
		} else if sortType == size {
			d := selected[j].Size() - selected[i].Size()
			if d > 0 {
				v = true
			} else if d == 0 {
				same = true
			}
			v = !v
		} else {
			// strcoll?
			v = strcmpi(selected[i].path, selected[j].path) == -1
		}
		if same {
			v = strcmpi(selected[i].path, selected[j].path) == -1
		}

		if reverseSort {
			v = !v
		}
		return
	})

	padding := 2
	smallestWord := 1
	var cols int
	var colWidths []int

	if longList {
		cols = 4
		colWidths = make([]int, cols)
		for _, v := range selected {
			li := getLongInfo(v)
			if decimalLen(int64(li.hardLinks)) > colWidths[0] {
				colWidths[0] = decimalLen(int64(li.hardLinks))
			}
			if len(li.userName) > colWidths[1] {
				colWidths[1] = len(li.userName)
			}
			if len(li.groupName) > colWidths[2] {
				colWidths[2] = len(li.groupName)
			}
			if humanReadable {
				if len(human(v.Size())) > colWidths[3] {
					colWidths[3] = len(human(v.Size()))
				}
			} else {
				if decimalLen(v.Size()) > colWidths[3] {
					colWidths[3] = decimalLen(v.Size())
				}
			}
		}
	} else {
		cols = width / (padding + smallestWord)
		colWidths = make([]int, cols)
		A:
		for {
			colWidths = colWidths[:cols]
			for i := range colWidths {
				colWidths[i] = 0
			}
			pos := (cols - 1) * padding
			for i, v := range selected {
				p := i % cols
				if len(v.path) > colWidths[p] {
					pos += len(v.path) - colWidths[p]
					if pos > width  {
						cols--
						if cols == 0 {
							cols = 1
							break A
						}
						continue A
					}
					colWidths[p] = len(v.path)
				}
			}
			break
		}
	}

	for i, v := range selected {
		if longList {
			li := getLongInfo(v)
			var timeStr string
			if now.Year() == v.ModTime().Year() {
				timeStr = v.ModTime().Format("Jan _2 15:04")
			} else {
				timeStr = v.ModTime().Format("Jan _2  2006")
			}
			linkPad := strings.Repeat(" ", colWidths[0] - decimalLen(int64(li.hardLinks)))
			userPad := strings.Repeat(" ", colWidths[1] - len(li.userName))
			groupPad := strings.Repeat(" ", colWidths[2] - len(li.groupName))
			name := v.path
			if v.Mode() & os.ModeSymlink != 0 {
				if l, err  := os.Readlink(root + v.path); err == nil {
					name = name + " -> " + l
				} else {
					log.Print(err)
				}
			}
			var sizeStr string
			if humanReadable {
				sizeStr = human(v.Size())
			} else {
				sizeStr = fmt.Sprintf("%d", v.Size())
			}

			sizePad := strings.Repeat(" ", colWidths[3] - len(sizeStr))

			fmt.Printf("%s %s%d %s%s %s%s %s%s %s %s\n", modeString(v.Mode()) , linkPad,
				li.hardLinks, li.userName, userPad, li.groupName, groupPad, sizePad, sizeStr, timeStr, name)
		} else {
			w := colWidths[i % cols]
			if i % cols == 0 {
				if i != 0 {
					fmt.Println()
				}
			}
			fmt.Printf("%s", v.path)
			if i % cols != cols -1 {
				fmt.Print(strings.Repeat(" ", (w - len(v.path)) + padding))
			}
		}
	}
	if !longList {
		fmt.Println()
	}
}

func main() {
	files := list.NewSliceList(&list.StringSlice{Data:os.Args}).(*list.StringSlice)
	options := list.NewSliceList(&list.StringSlice{}).(*list.StringSlice)

	files.Remove(0)
	for iter := files.Iterator(0); iter.Next(); {
		if v := files.Data[iter.Pos()]; strings.HasPrefix(v, "-") {
			options.Data[options.Append()] = v
			iter.Remove()
			if v == "--" {
				break
			}
		}
	}

	if files.Len() == 0 {
		files.Data[files.Append()] = "."
	}

	for iter := options.Iterator(0); iter.Next(); {
		if option := options.Data[iter.Pos()]; !strings.HasPrefix(option, "--") && len(option) > 2 {
			letters := list.NewSliceList(&list.ByteSlice{Data:[]byte(option[1:])}).(*list.ByteSlice)
			var removed bool
			for iter2 := letters.Iterator(letters.Len() - 1); iter2.Prev(); {
				options.Data[iter.Insert()] = "-" + string(letters.Data[iter2.Pos()])
				if !removed {
					iter.Remove()
					removed = true
				}
				iter.Prev()
			}
		}

		switch options.Data[iter.Pos()] {
		case "-d":
			showDirEntries = true
		case "-a":
			showAll = true
		case "-A":
			showAlmostAll = true
			showAll = true
		case "-t":
			sortType = modTime
		case "-S":
			sortType = size
		case "-r":
			reverseSort = true
		case "-l":
			longList = true
		case "-h":
			humanReadable = true
		case "-R":
			recursiveList = true
		case "-O":
			onlyHidden = true
		case "--help":
			fmt.Print(helpStr)
			os.Exit(0)
		default:
			log.Fatalf("unkown option %s", options.Data[iter.Pos()])
		}
	}

	if w, _, err := getTermSize(); err == nil {
		width = w
	} else {
		width = 80
	}

	selected := list.NewSliceList(&DisplayEntryList{}).(*DisplayEntryList)

	for iter := files.Iterator(0); iter.Next(); {
		fileName := files.Data[iter.Pos()]
		if showDirEntries {
			if stat, err := os.Lstat(fileName); err == nil {
				selected.Data[selected.Append()] = DisplayEntry{fileName, stat}
			} else {
				log.Print(err)
			}
			iter.Remove()
		} else {
			if stat, err := os.Lstat(fileName); err == nil {
				if stat.IsDir() {
					continue
				} else {
					selected.Data[selected.Append()] = DisplayEntry{fileName, stat}
					iter.Remove()
				}
			} else {
				log.Print(err)
				iter.Remove()
			}
		}
	}

	if selected.Len() > 0 && !recursiveList {
		display(selected.Data, "")
	}

	// directories
	for iter := files.Iterator(0); iter.Next(); {
		fileName := files.Data[iter.Pos()]

		if !recursiveList {
			if selected.Len() > 0 {
				selected.Clear()
				fmt.Println()
				fmt.Printf("%s:\n", fileName)
			} else if files.Len() > 1 {
				fmt.Printf("%s:\n", fileName)
			}
		}

		var total int64 = 0
		if file, err := os.Open(fileName); err == nil {
			if fileInfos, err := file.Readdir(0); err == nil {
				if showAll && !showAlmostAll && !recursiveList && !onlyHidden {
					if stat, err := os.Stat(fileName); err == nil {
						selected.Data[selected.Append()] = DisplayEntry{".", stat}
					} else {
						log.Print(err)
					}
					if parent, err := os.Stat(path.Clean(fileName + "/..")); err == nil {
						selected.Data[selected.Append()] = DisplayEntry{"..", parent}
					} else {
						log.Print(err)
					}
				}
				for _, v := range fileInfos {
					isHidden := strings.HasPrefix(v.Name(), ".")
					if !onlyHidden && (showAll || !isHidden) || onlyHidden && isHidden {
						total += v.Size()
						if recursiveList {
							path := path.Clean(fileName + "/" + v.Name())
							selected.Data[selected.Append()] = DisplayEntry{path, v}
							if v.IsDir() {
								files.Data[files.Append()] = path
							}
						} else {
							selected.Data[selected.Append()] = DisplayEntry{v.Name(), v}
						}
					}
				}
			} else {
				log.Print(err)
			}
			file.Close()
		} else {
			log.Print(err)
		}

		if longList && !recursiveList {
			if humanReadable {
				fmt.Printf("total %s\n", human(total))
			} else {
				fmt.Printf("total %d\n", total/1024)
			}
		}

		if !recursiveList {
			display(selected.Data, fileName + "/")
		}
	}

	if recursiveList && selected.Len() > 0 {
		log.Printf("sorting/displaying")
		display(selected.Data, "")
	}
}

var helpStr =
`Usage: ls [OPTION]... [FILE]...
List information about the FILEs (the current directory by default).
Sort entries alphabetically unless a sort option is given.
	-a					do not ignore entries starting with .
	-A					do not list implied . and ..
	-d					list directory entries instead of contents
	-t					sort by modification time, newest first
	-S					sort by file size
	-r					reverse order while sorting
	-l					use a long listing format
	-h					with -l, print sizes in human readable format
	-R					list subdirectories recursively, sorting all files
	-O					only list entries starting with .
	--help				display this help and exit
`