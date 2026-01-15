package exporter

import (
	"fmt"
	"os"
)

// TextWriter writes chunks to a text file.
type TextWriter struct {
	OutputPath string
}

// Write writes the chunks to a text file.
func (w *TextWriter) Write(root Root) error {
	outputContent := ProcessChunks(root)

	if err := os.WriteFile(w.OutputPath, []byte(outputContent), 0644); err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}

	return nil
}
