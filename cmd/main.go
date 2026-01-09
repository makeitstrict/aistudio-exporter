package main

import (
	"fmt"
	"os"

	"aistudio-exporter/internal/exporter"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aistudio-exporter",
	Short: "Extracts text chunks from a JSON file into a single text document",
}

var exportCmd = &cobra.Command{
	Use:   "export [input.json] [output.txt]",
	Short: "Exports text chunks from JSON to a text file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		output := args[1]
		if err := exporter.ExportChunks(input, output); err != nil {
			return err
		}
		fmt.Printf("Successfully exported from %s to %s\n", input, output)
		return nil
	},
}

func main() {
	rootCmd.AddCommand(exportCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
