# aistudio-exporter

A Go CLI utility for extracting text chunks from a JSON file (e.g., example.json) and saving them to a text file, separating each chunk with a "\n---\n" string.

## Installation

```bash
go mod tidy
```

## Build

```bash
go build -o aistudio-exporter ./cmd
```

## Usage

```bash
./aistudio-exporter export example.json output.txt
```

- Extracts only those chunks where `isThought != true`.
- Each chunk is separated by a `\n---\n` string in the resulting file.

## Testing

To run tests, execute:

```bash
go test ./...
```

## Dependencies
- [cobra](https://github.com/spf13/cobra) — for CLI
