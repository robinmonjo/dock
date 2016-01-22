# dock

`dock` is a "micro init system" for containers

#TODO

- finish system package
- add tests on system package --> add a vagrant file for this or it will be a pain in the asshole
- use process.wait to handle process end + manage signals in a goroutine
- finish testing and stuff

#TODO:

- port binding
- log rotate (inside the container or in bindmounted dir)
- stdout / stderr redirection
- stdout prefix + coloring
- check if err on start is a wait status and if so, return the exit code (i.e: 127 path not found)

Test TODO
- web-hook
