# dock

`dock` is a "micro init system" for linux containers

## Installation

````bash
curl -sL https://github.com/robinmonjo/dock/releases/download/v0.8/dock-v0.8.tgz | tar -C /usr/local/bin -zxf -
````

This will place the latest `dock` binary in `/usr/local/bin`

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

`dock` is written in Go and has no dependency. The binary can simply be added into a linux container image. It also provides some useful features :

- can call a web hook when process state changes (`starting`, `running` and `crashed`)
- can say a process started only when a given port is bound (if you start a web server, you may want to know when this one is ready to accept connections). Think container rotation during a deployment process
- smart stdin / stdout (see the `--io` flag for more information)
- can provide log rotation (see `--log-rotate` flag for more information)
- authoritarian signal transmission (see `--thug` flag for more information)

Note: `dock` may be used outside of a container, directly on a linux system

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

If specified, `dock` performs a HTTP PUT request with a JSON payload that contains information about the process and its environment:

````json
{
  "ps": {
    "status": "running",
    "net_interfaces": [
      {
        "name": "lo",
        "ipv4": "127.0.0.1",
        "ipv6": "::1"
      },
      {
        "name":"eth0",
        "ipv4":"172.17.0.2",
        "ipv6":"fe80::42:acff:fe11:2"
      }
    ]
  }
}
````

where `status` may be: `starting`, `running` or `crashed`. Note that if `--bind-port` flag is used, the `running` status is sent only once the given port is bound by one of `dock` children processes.

This payload will evolve to carry more useful information in the future.

#### `--bind-port`

Port `dock`'s child process is expected to bind. Port may be bound by any processes in the container. See `--strict-port-binding` for more control.

#### `--strict-port-binding`

If `--bind-port` is specified, this flag will ensure that the process is considered running only if the binder is a descendant process of `dock`. This is not really useful in container environment since dock will have PID 1 (hence any port in the container will be bound by a descendant). Be careful while using this flag (TODO: explain why)


#### `--log-rotate`

If given `--io` is a file, specifying `--log-rotate X` perform a log rotation every X hours:

- archive (gzip) the current log file by prepending a timestamp (in the stdout file directory)
- empty the current log file
- keep at most 5 log archives

#### `--stdout-prefix`

Add a prefix to stdout lines. Format: `prefix[:<color>]` where color may be white, green, blue, magenta, yellow, cyan or red

#### `--thug`

When you stop a docker container, a SIGTERM is sent to the process running it. The docker daemon then wait for a certain delay and if the container still exists, it will kill the process (using SIGKILL). This happens a lot with process ran as PID 1 inside a linux container, since the [kernel will treat PID 1 specially](http://lwn.net/Articles/532748/). 

With `dock` this behavior happens less frequently since `dock` runs as PID 1. However some program block or ignore some signals. For example, `sh` ignores the SIGTERM signal. Using `docker stop` on a container running `sh` will force the docker engine to kill the process. This may be frustrating.

The `--thug` flag allows to translate a stopping signal (SIGINT, SIGQUIT, SIGTERM) into a SIGKILL **if** the stopping signal is blocked or ignored by `dock`'s child process

## Working on `dock`

- use the Makefile and Dockerfile :)
- use the `-d` flag for verbosity
- the `utils.go` file contains some nice function to inspect and debug

# TODOs

- [why this ?](https://github.com/gliderlabs/docker-alpine/issues/143)
- more tests !

## Known issue (not blocking)

- process is not stopped if interactive and over the network and the connection is closed

