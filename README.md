# LS

## Install
``` bash
$ git clone https://github.com/timob/ls.git
Cloning into 'ls'...
remote: Enumerating objects: 210, done.
remote: Total 210 (delta 0), reused 0 (delta 0), pack-reused 210
Receiving objects: 100% (210/210), 1.45 MiB | 17.05 MiB/s, done.
Resolving deltas: 100% (87/87), done.
$ cd ls
$ go get
go get: added github.com/bradfitz/slice v0.0.0-20180809154707-2b758aa73013
go get: added github.com/daviddengcn/go-colortext v1.0.0
go get: added github.com/dustin/go-humanize v1.0.0
go get: added github.com/timob/ls v0.0.0-20171116232057-a724f1d86305
go get: added github.com/timob/sindex v0.0.0-20201206080312-1eedde862709
go get: added go4.org v0.0.0-20201209231011-d4a079459e60

# Make sure $GOBIN is set, to where the binary is placed.
$ go install
```

## Changes
* 2022-04-15: Add go.mod, v1.0.0
* 2017-11-03: Add --height option, with less allows for wide output
* 2016-11-26: Add -P path only mode option, add --width option, OSX support
* 2016-8-25: Add humanized timestamps with -h
* 2016-2-28: Fix user name, group name lookup on Unix
* 2016-2-27: Added color output

## Description
List directory Unix utility written in Go features:
* Works on Windows/Unix/OSX
* Same as GNU ls
* -R has changed, meaning recursively lists all files with relative path names as one group. This means sort options such as by time (-t) can show most recent file modified under a path. (Useful for example to see latest logs in /var/log)
* -O only list entries starting with.

```
Usage: ls [OPTION]... [FILE]...
List information about the FILEs (the current directory by default).
Sort entries alphabetically unless a sort option is given.
        -a                                      do not ignore entries starting with .
        -A                                      do not list implied . and ..
        -d                                      list directory entries instead of contents
        -t                                      sort by modification time, newest first
        -S                                      sort by file size
        -r                                      reverse order while sorting
        -l                                      use a long listing format
        -h                                      with -l, print sizes, time stamps in human readable format
        -R                                      list subdirectories recursively, sorting all files
        -P                                      when used with -R, enables path mode, only file paths are displayed
        -O                                      only list entries starting with .
        -C                                      list entries by columns
        -x                                      list entries by lines instead of by columns
        -1                                      list one file per line
        -i, --inode                             print the index number of each file
        --width=COLS                            assume screen width
        --height=LINES                          constrain column height to LINES
        --color[=WHEN]                          colorize the output WHEN defaults to 'always'
                                                or can be "never" or "auto".
        --use-c-strcoll                         use strcoll by making C call from Go when sorting file names
                                                instead of native string comparison function
        --help                                  display this help and exit
````

## Why?
This uses the SIndex https://github.com/timob/sindex slice indexing library to handle lists of options, file arguments, directory
lists. So really a use case for that library. IMHO it makes programming lists using iterators, insert, deleting and appending much
easier.

## Todo
* Add other GNU ls options.
* Fix terminal detection.
