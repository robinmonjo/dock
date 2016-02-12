# dock

`dock` is a "micro init system" for containers

## How signals are handled

`dock` acts as the PID 1 in the container. It forwards every signals to it's child process. When a child process die, `dock` will cleanly exists :

1. detect one of its child died
2. send SIGTERM to all processes remaining in its process tree
3. call `wait4` until no more children exist

Step 3 may block if some processes do not respond to the SIGTERM in step 2. If this is the case, a SIGKILL is sent after the SIGTERM (with a 5 seconds timeout)

## Inspirations / motivations

- [Yelp dumn-init](https://github.com/Yelp/dumb-init)
- [Phusion baseimage docker](https://github.com/phusion/baseimage-docker)

## Why ?

- detect when a port is bound (a container that start doesn't mean that your web server is ready to accept connections)
- log rotation may be useful
- stdin / stdout redirection can be super useful (especially over the network)
- web hooks are handy

#TODO

- test port binding check (to continue with nc ...)
- log rotations + test (take it from psdock3)
- stdout, stderr, stdin redirection
- stdout prefix + coloring
- check if err on start is a wait status and if so, return the exit code (i.e: 127 path not found)

- more testing

