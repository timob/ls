package main

import (
	"os"
	"github.com/timob/list"
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"
)

type DisplayEntry struct {
	path string
	os.FileInfo
}

type DisplayEntryList struct {
	Data []DisplayEntry
	list.Slice
}

func getTermSize() (int, int, error) {
	var dimensions [4]uint16

	fd := os.Stdout.Fd()
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}

func main() {
	files := list.NewSliceList(&list.StringSlice{Data:os.Args}).(*list.StringSlice)
	options := list.NewSliceList(&list.StringSlice{}).(*list.StringSlice)

	files.Remove(0)
	for iter := files.Iterator(0); iter.Next(); {
		if v := files.Data[iter.Pos()]; strings.HasPrefix(v, "-") {
			options.Data[options.Append()] = v
			iter.Remove()
			if strings.HasPrefix(v, "--") {
				break
			}
		}
	}

	var showDirEntries bool
	for iter := options.Iterator(0); iter.Next(); {
		switch options.Data[iter.Pos()] {
		case "-d":
			showDirEntries = true
		}
	}

	var width int
	if w, _, err := getTermSize(); err == nil {
		width = w
	} else {
		width = 80
	}

	selected := list.NewSliceList(&DisplayEntryList{}).(*DisplayEntryList)
	for iter := files.Iterator(0); iter.Next(); {
		if fileName := files.Data[iter.Pos()]; showDirEntries {
			if stat, err := os.Lstat(fileName); err == nil {
				selected.Data[selected.Append()] = DisplayEntry{fileName, stat}
			} else {
				log.Print(err)
			}
		} else {
			if stat, err := os.Stat(fileName); err == nil {
				if stat.IsDir() {
					if file, err := os.Open(fileName); err == nil {
						if fileInfos, err := file.Readdir(0); err == nil {
							for _, v := range fileInfos {
								selected.Data[selected.Append()] = DisplayEntry{v.Name(), v}
							}
						} else {
							log.Print(err)
						}
					} else {
						log.Print(err)
					}
				}
			} else {
				log.Print(err)
			}
		}

		padding := 2
		smallestWord := 1
		cols := width / (padding + smallestWord)
		colWidths := make([]int, cols)
A:
		for cols > 1 {
			colWidths = colWidths[:cols]
			for i := range colWidths {
				colWidths[i] = 0
			}
			pos := (cols - 1) * padding
			for i, v := range selected.Data {
				p := i % cols
				if len(v.path) > colWidths[p] {
					pos += len(v.path) - colWidths[p]
					if pos >= width {
						cols--
						continue A
					}
					colWidths[p] = len(v.path)
				}
			}
			break
		}

		for i, v := range selected.Data {
			w := colWidths[i % cols]
			if i % cols == 0 {
				if i != 0 {
					fmt.Println()
				}
			}
			fmt.Printf("%s", v.path)
			fmt.Print(strings.Repeat(" ", (w - len(v.path)) + padding))
		}

		fmt.Println()
		selected.Clear()
	}
}