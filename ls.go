package main

import (
	"os"
	"github.com/timob/list"
	"fmt"
	"log"
	"strings"
	"strconv"
)

type DisplayEntry struct {
	path string
	os.FileInfo
}

type DisplayEntryList struct {
	Data []DisplayEntry
	list.Slice
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
	if i, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil {
		width = i
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
		for cols > 1 {
			colWidths = colWidths[:cols]
			for i := range colWidths {
				colWidths[i] = 0
			}
			for i, v := range selected.Data {
				p := i % cols
				if len(v.path) > colWidths[p] {
					colWidths[p] = len(v.path)
				}
			}
			pos := 0
			for _, v := range colWidths {
				pos += v
			}
			pos += (cols - 1) * padding
			if pos >= width {
				cols--
			} else {
				break
			}
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