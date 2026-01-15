package main

import (
	"fmt"
	"os"
	"strings"

	"aistudio-exporter/internal/exporter"

	"github.com/spf13/cobra"
)

var (
	format string
)

var rootCmd = &cobra.Command{
	Use:   "aistudio-exporter",
	Short: "Extracts text chunks from a JSON file into a single text document or database",
}

var exportCmd = &cobra.Command{
	Use:   "export [input.json] [output]",
	Short: "Exports text chunks from JSON to a text file or SQLite database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		output := args[1]

		var writer exporter.Writer
		switch strings.ToLower(format) {
		case "txt", "text":
			writer = &exporter.TextWriter{OutputPath: output}
		case "sqlite", "db":
			writer = &exporter.SQLiteWriter{DBPath: output}
		default:
			return fmt.Errorf("unsupported format: %s (supported: txt, sqlite)", format)
		}

		if err := exporter.ExportChunks(input, writer); err != nil {
			return err
		}
		fmt.Printf("Successfully exported from %s to %s (format: %s)\n", input, output, format)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&format, "format", "f", "txt", "Output format: txt or sqlite")
}

func main() {
	rootCmd.AddCommand(exportCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
