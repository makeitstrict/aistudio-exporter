# aistudio-exporter

A Go CLI utility for extracting text chunks from a JSON file (e.g., example.json) and saving them to a text file or SQLite database, separating each chunk with a "\n---\n" string in text format.

## Installation

```bash
go mod tidy
```

## Build

```bash
go build -o aistudio-exporter ./cmd
```

## Usage

### Export to text file (default)

```bash
./aistudio-exporter export example.json output.txt
```

### Export to SQLite database

```bash
./aistudio-exporter export example.json output.db -f sqlite
```

or

```bash
./aistudio-exporter export example.json output.db --format sqlite
```

- Extracts only those chunks where `isThought != true`.
- Text format: Each chunk is separated by a `\n---\n` string in the resulting file.
- SQLite format: Chunks are stored in a `chunk_records` table with `id` and `text` columns.

## Testing

To run tests, execute:

```bash
go test ./...
```

## Dependencies
- [cobra](https://github.com/spf13/cobra) — for CLI
- [gorm](https://gorm.io/) — for SQLite database operations
