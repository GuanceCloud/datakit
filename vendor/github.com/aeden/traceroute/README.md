# Traceroute in Go

A traceroute library written in Go.

Must be run as sudo on OS X (and others)?

## CLI App

```sh
go build cmd/gotraceroute
sudo ./gotraceroute example.com
```

## Library

See the code in cmd/gotraceroute.go for an example of how to use the library from within your application.

The traceroute.Traceroute() function accepts a domain name and an options struct and returns a TracerouteResult struct that holds an array of TracerouteHop structs.

## Resources

Useful resources:

* http://en.wikipedia.org/wiki/Traceroute
* http://tools.ietf.org/html/rfc792
* http://en.wikipedia.org/wiki/Internet_Control_Message_Protocol

## Notes

* https://code.google.com/p/go/source/browse/src/pkg/net/ipraw_test.go
* http://godoc.org/code.google.com/p/go.net/ipv4
* http://golang.org/pkg/syscall/
