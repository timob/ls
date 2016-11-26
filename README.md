LS
========

####Changes
* 2016-11-26: Add -P path only mode option, add --width option, OSX support
* 2016-8-25: Add humanized timestamps with -h
* 2016-2-28: Fix user name, group name lookup on Unix
* 2016-2-27: Added color output

####Description
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
        --color[=WHEN]                          colorize the output WHEN defaults to 'always'
                                                or can be "never" or "auto".
        --use-c-strcoll                         use strcoll by making C call from Go when sorting file names
                                                instead of native string comparison function
        --help                          display this help and exit
````

####Cool usage
Add alias `alias ls='ls --color=always -h -C -P --width=$COLUMNS'`

This is good for intertactive usage, eg. `ls -R |less` will display all file paths below current directory in a wide format using
all of your terminal width. If there are long path names use very big width eg. `--width=10000`, this will allow you to pan around
in less (remeber to add the `-S` option to `$LESS`).

####Why?
This uses the SIndex https://github.com/timob/sindex slice indexing library to handle lists of options, file arguments, directory
lists. So really a use case for that library. IMHO it makes programming lists using iterators, insert, deleting and appending much
easier.

####Todo
* Add other GNU ls options
