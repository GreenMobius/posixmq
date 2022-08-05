# posixmq

PosixMQ is a golang implementation of posix messaging queues using the unix syscall interface.

As POSIX message queues only exist on POSIX, there is obviously limited platform support. The package will only be maintained for Debian and Ubuntu.

The project will be able to be built and tested in a standard debian container using scuba and docker.

See [the scuba project](https://github.com/JonathonReinhart/scuba) for documentation on how to install and use the scuba tool.
