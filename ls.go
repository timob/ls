package main

import (
	"fmt"
	"github.com/bradfitz/slice"
	ct "github.com/daviddengcn/go-colortext"
	. "github.com/timob/ls/lib"
	"github.com/timob/sindex"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"github.com/dustin/go-humanize"
)

type DisplayEntry struct {
	path string
	os.FileInfo
}

type DisplayEntryList struct {
	Data []DisplayEntry
	sindex.List
}

var now = time.Now()

var showDirEntries bool
var showAll bool
var showAlmostAll bool
var longList bool

const (
	name    int = iota
	modTime int = iota
	size    int = iota
)

var sortType int = name
var reverseSort bool
var humanReadable bool
var recursiveList bool
var onlyHidden bool
var width int
var oneColumn bool
var listBylines bool
var useCstrcoll bool
var showInode bool
var pathMode bool
var height int
var wide bool

type colorDef struct {
	fg, bg byte
	bright bool
}

var fileColors map[string]colorDef

var useColor bool

var colorSet bool

func setColor(def colorDef) {
	colorSet = true
	ct.ChangeColor(ct.Color(def.fg), def.bright, ct.Color(def.bg), false)
}

func resetColor() {
	if colorSet {
		ct.ResetColor()
		colorSet = false
	}
}

func setColorForFile(info os.FileInfo) {
	mode := info.Mode()
	var fileType string
	if mode&os.ModeDir != 0 {
		if mode&os.ModeSticky != 0 {
			if mode&(1<<1) != 0 {
				fileType = "tw"
			} else {
				fileType = "st"
			}
		} else if mode&(1<<1) != 0 {
			fileType = "ow"
		} else {
			fileType = "di"
		}
	} else if mode&os.ModeSymlink != 0 {
		fileType = "ln"
	} else if mode&os.ModeNamedPipe != 0 {
		fileType = "pi"
	} else if mode&os.ModeSocket != 0 {
		fileType = "so"
	} else if mode&os.ModeDevice != 0 {
		fileType = "bd"
	} else if mode&os.ModeCharDevice != 0 {
		fileType = "cd"
	} else if mode&os.ModeSetuid != 0 {
		fileType = "su"
	} else if mode&os.ModeSetgid != 0 {
		fileType = "sg"
	} else if mode&(1<<6|1<<3|1) != 0 {
		fileType = "ex"
	} else {
		name := info.Name()
		if n := strings.LastIndex(name, "."); n != -1 && n != len(name)-1 {
			key := "*" + name[n:]
			if _, ok := fileColors[key]; ok {
				fileType = key
			}
		}
	}
	if fileType != "" {
		setColor(fileColors[fileType])
	}
}

func human(n int64) string {
	var i int64
	var w = n
	var uSize int64 = 1
	for i = 0; i < 12; i++ {
		if w/1024 == 0 {
			break
		}
		w = w / 1024
		uSize *= 1024
	}

	var f int64
	var unit string
	f = (n - w*uSize)
	if f != 0 && i > 0 {
		if w < 10 {
			if f != 0 {
				lowerSize := uSize / 1024
				// magic plus one here (seems to be what GNU ls does)
				tenth := int64(1024/10) + 1
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
		if n/10 == 0 {
			break
		}
		n = n / 10
	}
	return
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
	if mode&os.ModeDir != 0 {
		output[0] = 'd'
	} else if mode&os.ModeSymlink != 0 {
		output[0] = 'l'
	} else if mode&os.ModeNamedPipe != 0 {
		output[0] = 'p'
	} else if mode&os.ModeSocket != 0 {
		output[0] = 's'
	} else if mode&os.ModeCharDevice != 0 && mode&os.ModeDevice != 0 {
		output[0] = 'c'
	}

	const rwx = "rwxrwxrwx"
	for i, c := range rwx {
		bitSet := mode&(1<<uint(9-1-i)) != 0
		if bitSet {
			if (i == 2 && mode&os.ModeSetuid != 0) || (i == 5 && mode&os.ModeSetgid != 0) {
				output[i+1] = 's'
			} else if i == 8 && mode&os.ModeSticky != 0 {
				output[i+1] = 't'
			} else {
				output[i+1] = byte(c)
			}
		} else if (i == 2 && mode&os.ModeSetuid != 0) || (i == 5 && mode&os.ModeSetgid != 0) {
			output[i+1] = 'S'
		} else if i == 8 && mode&os.ModeSticky != 0 {
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
				same = selected[i].ModTime().Equal(selected[j].ModTime())
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
			if !useCstrcoll {
				v = strcmpi(selected[i].path, selected[j].path) == -1
			} else {
				v = Strcoll(selected[i].path, selected[j].path) < 0
			}
		}
		if same {
			if !useCstrcoll {
				v = strcmpi(selected[i].path, selected[j].path) == -1
			} else {
				v = Strcoll(selected[i].path, selected[j].path) < 0
			}
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
		cols = 6
		colWidths = make([]int, cols)
		for _, v := range selected {
			li := GetLongInfo(v)
			if decimalLen(int64(li.HardLinks)) > colWidths[0] {
				colWidths[0] = decimalLen(int64(li.HardLinks))
			}
			if len(li.UserName) > colWidths[1] {
				colWidths[1] = len(li.UserName)
			}
			if len(li.GroupName) > colWidths[2] {
				colWidths[2] = len(li.GroupName)
			}
			if humanReadable {
				if len(human(v.Size())) > colWidths[3] {
					colWidths[3] = len(human(v.Size()))
				}
				if len(humanize.Time(v.ModTime())) > colWidths[4] {
					colWidths[4] = len(humanize.Time(v.ModTime()))
				}
			} else {
				if decimalLen(v.Size()) > colWidths[3] {
					colWidths[3] = decimalLen(v.Size())
				}
			}
			if showInode {
				dl := decimalLen(int64(li.Ino))
				if dl > colWidths[5] {
					colWidths[5] = dl
				}
			}
		}
	} else {
		if oneColumn {
			cols = 1
		} else if wide {
			colHeight := height - 1
			cols = len(selected) / colHeight
			if len(selected) % colHeight != 0 {
				cols++
			}
		} else {
			cols = width / (padding + smallestWord)
		}
		colWidths = make([]int, cols)
	A:
		for {
			colWidths = colWidths[:cols]
			for i := range colWidths {
				colWidths[i] = 0
			}
			pos := (cols - 1) * padding
			for i := range selected {
				p := i % cols
				var j int
				if listBylines {
					j = i
				} else {
					var per int
					if len(selected) % cols == 0 {
						per = len(selected) / cols
					} else {
						per = len(selected) / cols + 1
					}
					square := per * cols
					if len(selected) <= square - per {
						cols--
						if cols == 0 {
							cols = 1
							break A
						}
						continue A
					}
					// if needed skip empty rows in last column
					// lastFullRow is index of last row with all cols present
					lastFullRow := (len(selected) - 1) % per
					curRow := i / cols
					if curRow > lastFullRow  {
						diff := (i - (lastFullRow + 1) * cols)
						p = diff % (cols - 1)
						curRow = lastFullRow + 1 + diff / (cols - 1)
					}
					j = (per * p) + curRow
				}
				v := selected[j]
				l := len(v.path)
				if showInode {
					li := GetLongInfo(v)
					l += decimalLen(int64(li.Ino)) + 1
				}
				if l > colWidths[p] {
					pos += l - colWidths[p]
					if pos > width && !wide {
						cols--
						if cols == 0 {
							cols = 1
							break A
						}
						continue A
					}
					colWidths[p] = l
				}
			}
			break
		}
	}

	for i := range selected {
		var j int
		adjCols := cols
		p := i % cols
		if listBylines || longList {
			j = i
		} else {
			var per int
			if len(selected) % cols == 0 {
				per = len(selected) / cols
			} else {
				per = len(selected) / cols + 1
			}
			lastFullRow := (len(selected) - 1) % per
			curRow := i / cols
			if curRow > lastFullRow {
				adjCols = cols - 1
				diff := (i - (lastFullRow + 1) * cols)
				p = diff % (cols - 1)
				curRow = lastFullRow + 1 + diff / (cols - 1)
			}
			j = (per * p) + curRow
		}
		v := selected[j]
		var linkTarget string
		var brokenLink bool
		var linkInfo os.FileInfo
		if v.Mode()&os.ModeSymlink != 0 {
			if l, err := os.Readlink(root + v.path); err == nil {
				linkTarget = l
				if i, err := os.Stat(root + v.path); err != nil {
					brokenLink = true
				} else {
					linkInfo = i
				}
			} else {
				log.Print(err)
			}
		}

		if longList {
			li := GetLongInfo(v)
			var timeStr string
			timePad := ""
			if humanReadable {
				timeStr = humanize.Time(v.ModTime())
				timePad = strings.Repeat(" ", colWidths[4]-len(timeStr))
			} else if now.Year() == v.ModTime().Year() {
				timeStr = v.ModTime().Format("Jan _2 15:04")
			} else {
				timeStr = v.ModTime().Format("Jan _2  2006")
			}
			linkPad := strings.Repeat(" ", colWidths[0]-decimalLen(int64(li.HardLinks)))
			userPad := strings.Repeat(" ", colWidths[1]-len(li.UserName))
			groupPad := strings.Repeat(" ", colWidths[2]-len(li.GroupName))
			var sizeStr string
			if humanReadable {
				sizeStr = human(v.Size())
			} else {
				sizeStr = fmt.Sprintf("%d", v.Size())
			}

			sizePad := strings.Repeat(" ", colWidths[3]-len(sizeStr))

			inodeStr := ""
			if showInode {
				inodeStr = strings.Repeat(" ", colWidths[5]-decimalLen(int64(li.Ino))) + fmt.Sprintf("%d ", li.Ino)
			}
			if useColor {
				fmt.Printf("%s%s %s%d %s%s %s%s %s%s %s%s ", inodeStr, modeString(v.Mode()), linkPad,
					li.HardLinks, li.UserName, userPad, li.GroupName, groupPad, sizePad, sizeStr, timePad, timeStr)
				if brokenLink {
					setColor(fileColors["or"])
				} else {
					setColorForFile(v.FileInfo)
				}
				fmt.Printf("%s", v.path)
				resetColor()
				if linkTarget != "" {
					fmt.Printf(" -> ")
					if brokenLink {
						setColor(fileColors["or"])
					} else {
						setColorForFile(linkInfo)
					}
					fmt.Printf("%s", linkTarget)
					resetColor()
				}
				fmt.Println()
			} else {
				name := v.path
				if v.Mode()&os.ModeSymlink != 0 {
					name = name + " -> " + linkTarget
				}
				fmt.Printf("%s%s %s%d %s%s %s%s %s%s %s%s %s\n", inodeStr, modeString(v.Mode()), linkPad,
					li.HardLinks, li.UserName, userPad, li.GroupName, groupPad, sizePad, sizeStr, timePad, timeStr, name)
			}
		} else {
			w := colWidths[p]
			if p == 0 {
				if i != 0 {
					fmt.Println()
				}
			}
			if useColor {
				if brokenLink {
					setColor(fileColors["or"])
				} else {
					setColorForFile(v.FileInfo)
				}
			}
			l := len(v.path)
			if showInode {
				li := GetLongInfo(v)
				l += decimalLen(int64(li.Ino)) + 1
				fmt.Printf("%d ", li.Ino)
			}
			fmt.Printf("%s", v.path)
			if useColor {
				resetColor()
			}
			if p != adjCols-1 {
				fmt.Print(strings.Repeat(" ", (w-l)+padding))
			}
		}
	}
	if !longList {
		fmt.Println()
	}
}

func main() {
	exit := 0
	files := sindex.InitListType(&sindex.StringList{Data: os.Args}).(*sindex.StringList)
	options := sindex.InitListType(&sindex.StringList{}).(*sindex.StringList)

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

	if !IsTerminal(1) {
		oneColumn = true
	}

	if w, h, err := GetTermSize(); err == nil {
		width = w
		height = h
	} else {
		width = 80
		height = 25
	}

	for iter := options.Iterator(0); iter.Next(); {
		if option := options.Data[iter.Pos()]; !strings.HasPrefix(option, "--") && len(option) > 2 {
			letters := sindex.InitListType(&sindex.ByteList{Data: []byte(option[1:])}).(*sindex.ByteList)
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

		var helpStr = `Usage: ls [OPTION]... [FILE]...
List information about the FILEs (the current directory by default).
Sort entries alphabetically unless a sort option is given.
	-a					do not ignore entries starting with .
	-A					do not list implied . and ..
	-d					list directory entries instead of contents
	-t					sort by modification time, newest first
	-S					sort by file size
	-r					reverse order while sorting
	-l					use a long listing format
	-h					with -l, print sizes, time stamps in human readable format
	-R					list subdirectories recursively, sorting all files
	-P					when used with -R, enables path mode, only file paths are displayed
	-O					only list entries starting with .
	-C					list entries by columns
	-x					list entries by lines instead of by columns
	-1					list one file per line
	-i, --inode				print the index number of each file
	--width=COLS				assume screen width
	--color[=WHEN]				colorize the output WHEN defaults to 'always'
						or can be "never" or "auto".
	--use-c-strcoll				use strcoll by making C call from Go when sorting file names
						instead of native string comparison function
	--help				display this help and exit
`
		option := options.Data[iter.Pos()]
		switch option {
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
		case "-P":
			pathMode = true
		case "-O":
			onlyHidden = true
		case "-x":
			listBylines = true
		case "-C":
			listBylines = false
			oneColumn = false
		case "-1":
			oneColumn = true
		case "--inode":
			fallthrough
		case "-i":
			showInode = true
		case "-W":
			wide = true
		case "--color":
			fallthrough
		case "--color=always":
			useColor = true
		case "--color=never":
			useColor = false
		case "--color=auto":
			if IsTerminal(1) {
				useColor = true
			} else {
				useColor = false
			}
		case "--use-c-strcoll":
			fallthrough
		case "--use-c-strcoll=yes":	
			useCstrcoll = true
		case "--use-c-strcoll=no":	
			useCstrcoll = false
		case "--help":
			fmt.Print(helpStr)
			os.Exit(0)
		default:
			if strings.HasPrefix(option, "--width=") {
				numStr := strings.TrimPrefix(option, "--width=")
				if _, err := fmt.Sscanf(numStr, "%d", &width); err != nil {
					log.Fatalf("invalid line width: %s", numStr)
				}
			} else if strings.HasPrefix(option, "--height=") {
				numStr := strings.TrimPrefix(option, "--height=")
				if _, err := fmt.Sscanf(numStr, "%d", &height); err != nil {
					log.Fatalf("invalid line width: %s", numStr)
				}
			} else {
				log.Fatalf("unkown option %s", option)
			}
		}
	}

	if useColor {
		colorBytesMap := map[string][]byte{
			"di": {1, 34},
			"ln": {1, 36},
			"pi": {40, 33},
			"so": {1, 35},
			"bd": {40, 33, 1},
			"cd": {40, 33, 1},
			"or": {40, 31},
			"su": {37, 41},
			"sg": {30, 43},
			"tw": {30, 42},
			"ow": {34, 42},
			"st": {37, 44},
			"ex": {01, 32},
		}
		lsColorsEnv := os.Getenv("LS_COLORS")
		colorDefs := strings.Split(lsColorsEnv, ":")
		for _, def := range colorDefs {
			tokens := strings.Split(def, "=")
			if len(tokens) != 2 {
				continue
			}
			colors := strings.Split(tokens[1], ";")
			var colorBytes []byte
			for _, color := range colors {
				if n, err := strconv.ParseInt(color, 10, 8); err == nil {
					colorBytes = append(colorBytes, byte(n))
				}
			}
			colorBytesMap[tokens[0]] = colorBytes
		}
		fileColors = make(map[string]colorDef)
		for k, v := range colorBytesMap {
			var bright bool
			var fg, bg int = 0, 0
			for _, b := range v {
				if b == 0 {
					bright = false
				} else if b == 1 {
					bright = true
				} else if b >= 30 && b < 38 {
					fg = int(b-30) + 1
				} else if b >= 40 && b < 48 {
					bg = int(b-40) + 1
				}
			}
			fileColors[k] = colorDef{byte(fg), byte(bg), bright}
		}
	}

	selected := sindex.InitListType(&DisplayEntryList{}).(*DisplayEntryList)

	for iter := files.Iterator(0); iter.Next(); {
		fileName := files.Data[iter.Pos()]
		if showDirEntries {
			if stat, err := os.Lstat(fileName); err == nil {
				selected.Data[selected.Append()] = DisplayEntry{fileName, stat}
			} else {
				log.Print(err)
				exit = 2
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
				exit = 2
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
			if names, err := file.Readdirnames(0); err == nil {
				for _, name := range names {
					isHidden := strings.HasPrefix(name, ".")
					if !onlyHidden && (showAll || !isHidden) || onlyHidden && isHidden {
						if v, err := os.Lstat(fileName + "/" + name); err == nil {
							total += v.Size()
							if recursiveList {
								path := path.Clean(fileName + "/" + v.Name())
								if !v.IsDir() || !pathMode {
									selected.Data[selected.Append()] = DisplayEntry{path, v}
								}
								if v.IsDir() {
									files.Data[files.Append()] = path
								}
							} else {
								selected.Data[selected.Append()] = DisplayEntry{v.Name(), v}
							}
						} else {
							log.Print(err)
							exit = 1
						}
					}
				}
			} else {
				log.Print(err)
				exit = 1
			}
			file.Close()
		} else {
			log.Print(err)
			exit = 1
		}

		if longList && !recursiveList {
			if humanReadable {
				fmt.Printf("total %s\n", human(total))
			} else {
				fmt.Printf("total %d\n", total/1024)
			}
		}

		if !recursiveList && selected.Len() > 0 {
			display(selected.Data, fileName+"/")
		}
	}

	if recursiveList && selected.Len() > 0 {
//		log.Printf("sorting/displaying")
		display(selected.Data, "")
	}
	os.Exit(exit)
}
