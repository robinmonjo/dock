# dock

`dock` is a "micro init system" for containers

## Inspirations / motivations

- [Yelp dumn-init](https://github.com/Yelp/dumb-init)
- [Phusion baseimage docker](https://github.com/phusion/baseimage-docker)

## How signals are handled

`dock` acts as the PID 1 in the container. It forwards every signals to it's child process. When a child process dies, `dock` :

1. detect one of its child died
2. send SIGTERM to all processes remaining in its process tree
3. call `wait4` until no more children exist

Step 3 may block if some processes do not respond to the SIGTERM in step 2. If this is the case, a SIGKILL is sent after the SIGTERM (within a 5 seconds timeout)

## Why `dock` ?

`dock` is written in Go and has 0 dependency. The binary can simply be added into a linux container image. It also provides some useful features :

- can call a web hook when process state changes (`starting`, `running` and `crashed`)
- can say a process started only when a given port is bound (if you start a web server, you may want to know when this one is ready to accept connections)
- smart stdin / stdout (see the `--io` flag for more information)
- can provide log rotation (see `--log-rotate` flag for more information)

## Usage

`dock [OPTIONS] command`

#### `--io`

Allows to redirect process stdin / stdout:

- `dock bash` run bash with current stdin and stdout
- `dock --io file:///var/run/process.log server.go` redirect stdout to the given file. Stdin stay unchanged
- `dock --io tcp://192.168.1.9:2567 bash` make stdin and stdout go over a tcp connection
- `dock --io tls://192.168.1.9:2657 bash` make stdin and stdout go over a tls connection

Every URL scheme supported by Go's `net.Dial` are supported by `dock`

#### `--web-hook`

If specified, `dock` performs a HTTP PUT request with the JSON payload:

````json
{
  "ps": {
    "status": "<status>"
  }
}
````

where `<status>` may be: `starting`, `running` or `crashed`. Note that if `--bind-port` flag is used, the `running` status is sent only once the given port is bound by one of `dock` children processes

#### `--bind-port`

Port `dock`'s child process is expected to bind. Port may be bound by a descendant of `dock` (not only its direct child). This flag is useful only when the `--web-hook` flag is used.

#### `--log-rotate`

If given `--io` is a file, specifying `-log-rotate X` perform a log rotation every X hours:

- archive (gzip) the current log file by prepending a timestamp (in the stdout file directory)
- empty the current log file
- keep at most 5 log archives

#### `--stdout-prefix`

Add a prefix to stdout lines. Format: `prefix[:<color>]` where color may be white, green, blue, magenta, yellow, cyan or red

## Working on `dock`

Use the Makefile and Dockerfile :)

# TODOs

- actually implement the kill timeout as it is documented :)
- check if err on start is a wait status and if so, return the exit code (i.e: 127 path not found)
- test port binding with `nc` and see [why this](https://github.com/gliderlabs/docker-alpine/issues/143) ?
- fix log rotate test that fails regularly
- more tests !

## Known issue (not blocking)

- stdout prefix + coloring only available over network but sometimes prefix get written at the end of a line
- process is not stopped if interactive and over the network and the connection is closed

