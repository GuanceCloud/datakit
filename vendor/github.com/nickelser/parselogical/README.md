# parselogical

Parses the output of the Postgres logical replication test_decoder.

## Install

To install, run:

```sh
go get github.com/nickelser/parselogical
```

This will give you the `github.com/nickelser/parselogical` library, assuming your Go paths are setup correctly.

## Usage

```go
ptd := parselogical.NewParseResult(walString)
ptd.Parse()

// now you can access the columns via
ptd.Columns["id"] // for example
```

TODO: more examples!

## Contributing

Everyone is encouraged to help improve this project. Here are a few ways you can help:

- [Report bugs](https://github.com/nickelser/parselogical/issues)
- Fix bugs and [submit pull requests](https://github.com/nickelser/parselogical/pulls)
- Write, clarify, or fix documentation
- Suggest or add new features
