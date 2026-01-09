package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ExportChunks reads input JSON file, extracts text chunks and saves them to output file.
func ExportChunks(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}

	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	outputContent := ProcessChunks(root)

	if err := os.WriteFile(outputPath, []byte(outputContent), 0644); err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}

	return nil
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
