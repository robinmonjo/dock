## system

Package system provide some information that could be retrieved by scanning the `profs` on linux.

It uses `netstat -n -l -p` to retrieve data about opened socket and their process owner

It uses `pgrep -P $PID` to retrieve children PID of a given process

It reads `/proc/$PID/status` to get back data on a process

`netstat` and `pgrep` are used because they are widely available, and retrieving information they provide "by hand" would have been hard and probably less efficient
