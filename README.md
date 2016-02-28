LS
========

####Changes
* 2016-2-28: Fix user name, group name lookup on Unix
* 2016-2-27: Added color output

####Description
List directory Unix utility written in Go features:
* Works on Windows/Unix
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
        -h                                      with -l, print sizes in human readable format
        -R                                      list subdirectories recursively, sorting all files
        -O                                      only list entries starting with .
        -1                                      list one file per line
        --color[=WHEN]                          colorize the output WHEN defaults to 'always'
                                                or can be "never" or "auto".
        --help                                  display this help and exit
````


####Why?
This uses the https://github.com/timob/sindex slice indexing library to handle lists of options, file arguments, directory lists.
So really a use case for that library. IMHO it makes programming lists using iterators, insert, deleting and appending much easier.

####Todo
* Add other GNU ls options
