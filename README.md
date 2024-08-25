# sus

[![codecov](https://codecov.io/github/quartzinquartz/sus/graph/badge.svg?token=3ONODB7RK5)](https://codecov.io/github/quartzinquartz/sus)

Inspired by `sort | uniq -c | sort -n` workflows, sus is a CLI line frequency analyzer.

## Features

- Read input from files or stdin (or both)
- Show most frequent or least frequent lines
- Support for percentage-based filtering
- Case-insensitive counting option
- Aggregate results across multiple inputs
- Sorts alphabetically on ties

## Installation

To install sus, use the following command:

```bash
go get github.com/quartzinquartz/sus
```

## Usage

Basic usage:

```bash
sus -high 5 -file input.txt
grep '123' /var/log/access.log | sus -high 5
grep '123' /var/log/access.log | sus -file $HOME/prepped_file.txt -high 5 --aggregate
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
