package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// maxInputSize limits untrusted exports so a single file cannot exhaust memory.
const maxInputSize int64 = 64 << 20

// Writer defines the interface for exporting chunks.
type Writer interface {
	Write(root Root) error
}

// ExportChunks reads input JSON file and saves chunks using the provided writer.
func ExportChunks(inputPath string, writer Writer) error {
	root, err := readAndParse(inputPath)
	if err != nil {
		return err
	}

	return writer.Write(root)
}

func readAndParse(inputPath string) (Root, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return Root{}, fmt.Errorf("error reading input file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxInputSize+1))
	if err != nil {
		return Root{}, fmt.Errorf("error reading input file: %w", err)
	}
	if int64(len(data)) > maxInputSize {
		return Root{}, fmt.Errorf("input file exceeds maximum size of %d bytes", maxInputSize)
	}

	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		return Root{}, fmt.Errorf("error parsing JSON: %w", err)
	}

	return root, nil
}

// ProcessChunks filters chunks and joins their text.
func ProcessChunks(root Root) string {
	var texts []string
	for _, chunk := range root.ChunkedPrompt.Chunks {
		if chunk.IsThought {
			continue
		}
		if chunk.Text != "" {
			texts = append(texts, chunk.Text)
		}
	}

	return strings.Join(texts, "\n---\n")
}
