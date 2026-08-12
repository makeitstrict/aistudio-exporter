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

	file, err := os.OpenFile(w.OutputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}
	defer file.Close()

	// Exported conversations can contain secrets. Tighten permissions even when
	// overwriting a file that was previously created with a permissive mode.
	if err := file.Chmod(0600); err != nil {
		return fmt.Errorf("error securing output file: %w", err)
	}
	if _, err := file.WriteString(outputContent); err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}

	return nil
}
