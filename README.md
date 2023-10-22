# Go Kea Stats Exporter

This tool serves [Kea](https://www.isc.org/kea/) statistics for consumption by
[Prometheus](https://prometheus.io/).

*Important:* this currently only supports the DHCPv4 server shipped with Kea. I
currently have no use for DHCPv6 and so have no easy way to test it, though I am
open to adding v6 support if:

- whoever contributes the code is willing to keep maintaining it
- there is some sort of test for the v6 portion of the code (unit tests, system
  tests etc)

## Usage

```
Usage of gkse:
  -c string
        if nonempty, load kea JSON config from file instead of querying unix domain socket
  -cl
        Enable color in logs (dault: false)
  -f string
        if nonempty, load stats JSON from file instead of querying unix domain socket
  -l string
        IP:port to listen on (default ":9988")
  -namespace string
        Namespace (prefix) to use for Prometheus metrics (default "kea")
  -s string
        Path to Kea control socket (default "/run/kea/kea4-ctrl-socket")
  -timeout duration
        Timeout for webserver reading client request (default 3s)
```

Note that typically, unprivileged users are not allowed to read drom/write to
the Kea control socket. On most distributions, the socket is owned by the Kea
system user/group. You can either run GKSE as that user, or run it as that group
(if you can figure out how change the default group Kea uses for the socket).

The exporter will not keep an open connection to the control socket, but instead
only open the socket if a request to the `/metrics` endpoint is made, issue the
necessary queries for stats and the config, and close the socket. As a result,
the exporter does not care whether Kea is runnning at startup, or if it is
restarted at a later point.
