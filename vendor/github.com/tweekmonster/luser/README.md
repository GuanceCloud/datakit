# luser

[![GoDoc](https://godoc.org/github.com/tweekmonster/luser?status.svg)](https://godoc.org/github.com/tweekmonster/luser)
[![Build Status](https://travis-ci.org/tweekmonster/luser.svg?branch=master)](https://travis-ci.org/tweekmonster/luser)

`luser` is a drop-in replacement for `os/user` which allows you to lookup users
and groups in cross-compiled builds without `cgo`.


## Overview

`os/user` requires `cgo` to lookup users using the target OS's API.  This is
the most reliable way to look up user and group information.  However,
cross-compiling means that `os/user` will only work for the OS you're using.
`user.Current()` is usable when building without `cgo`, but doesn't always
work.  The `$USER` and `$HOME` variables could be different from what you
expect or not even exist.

If you want to cross-compile a relatively simple program that needs to write a
config file somewhere in the user's directory, the last thing you want to do is
figure out some elaborate build scheme involving virtual machines.


## Usage

`luser` has the same API as `os/user`.  You should be able to just replace
`user.` with `luser.` in your files and let `goimports` do the rest.  The
returned `*User` and `*Group` types will have an `IsLuser` field indicating
whether or not a fallback method was used.


## Install

Install the package with:

```shell
$ go get github.com/tweekmonster/luser
```

A sample program called `luser` can be installed if you want to see the
fallback results on your platform:

```shell
$ CGO_ENABLED=0 go install github.com/tweekmonster/luser/cmd/luser
$ luser -c
$ luser username
$ luser 1000
```


## Fallback lookup methods

`os/user` functions are used when built with `cgo`.  Otherwise, it falls back
to one of the following lookup methods:

| Method        | Used for                                                       |
|---------------|----------------------------------------------------------------|
| `/etc/passwd` | Parsed to lookup user information. (Unix, Linux)               |
| `/etc/group`  | Parsed to lookup group information. (Unix, Linux)              |
| `getent`      | Optional. Find user/group information. (Unix, Linux)           |
| `dscacheutil` | Lookup user/group information via Directory Services. (Darwin) |
| `id`          | Finding a user's groups when using `GroupIds()`.               |

**Note:** Windows should always work regardless of the build platform since it
uses `syscall` instead of `cgo`.


## Caveats

- `luser.User` and `luser.Group` are new types.  The underlying `user.*` types
  are embedded, however (e.g. `u.User`).
- The lookup methods use `exec.Command()` and will be noticeably slower if
  you're looking up users and groups a lot.
- Group-releated functions will only be available when compiling with Go 1.7+.
