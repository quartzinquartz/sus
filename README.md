# sus

[![codecov](https://codecov.io/github/quartzinquartz/sus/graph/badge.svg?token=3ONODB7RK5)](https://codecov.io/github/quartzinquartz/sus)

Inspired by `sort | uniq -c | sort -n` workflows, sus is a CLI line frequency analyzer.

## Features

- Read input from files or stdin (or both)
- Show most frequent or least frequent lines
- Support for percentage-based filtering
- Case-insensitive counting option
- Aggregate results across multiple inputs
- Sorts lexicographically on ties

## Installation

Do the following to install sus from source:

1. Ensure you have Go installed on your system.
2. Run: `go install github.com/quartzinquartz/sus@latest`

This will install the `sus` binary in your `$GOPATH/bin` directory. Ensure this directory is in your system's PATH.

## Usage

Basic usage:

```bash
sus -high 5 -file input.txt
grep '123' /var/log/access.log | sus -high 5
grep '123' /var/log/access.log | sus -file $HOME/prepped_file.txt -high 5 -aggregate
```
These will show the 5 most frequent lines in the input.

For more options: `sus -help`

## Examples

1. Show 10 least frequent lines from stdin:
```
awk '/foo/' input.txt | sus -low 10
```
2. Show top 5% most frequent lines from multiple files:
```
sus -hp 5 -file file1.txt,file2.txt
```
3. Show aggregate results across all input sources:
```
sus -high 5 -low 2 -file file1.txt,file2.txt -aggregate
```

## Testing

To run the tests:
```
go test -v ./...
```

To see test coverage:
```
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Benchmarking

To run benchmarks:
```
go test -bench=. -benchmem
```

To reduce output on screen, you can either redirect to a file or use a filter like `go test -bench=. -benchmem | grep -E 'goos|goarch|cpu|Benchmark|github|allocs|PASS|FAIL'`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
