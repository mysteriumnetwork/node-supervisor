# node-supervisor

Supervisor is a system background service that allows seamless installation/running of [Mysterium node](https://github.com/mysteriumnetwork/node) under elevated permissions.
Clients (e.g. desktop applications) can ask the supervisor to RUN or KILL the node instance via OS dependent mechanism.

Currently, only macOS is supported and it is using unix domain sockets for communication.

For usage see:

```
myst_supervisor -help
```
