package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"aistudio-exporter/internal/exporter"

	"github.com/spf13/cobra"
)

func resetRootCmd() {
	rootCmd = &cobra.Command{
		Use:   "aistudio-exporter",
		Short: "Extracts text chunks from a JSON file into a single text document or database",
	}
	
	format = "txt" // reset format flag
	
	exportCmd = &cobra.Command{
		Use:   "export [input.json] [output]",
		Short: "Exports text chunks from JSON to a text file or SQLite database",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			output := args[1]

			var writer exporter.Writer
			switch format {
			case "txt", "text":
				writer = &exporter.TextWriter{OutputPath: output}
			case "sqlite", "db":
				writer = &exporter.SQLiteWriter{DBPath: output}
			default:
				return fmt.Errorf("unsupported format: %s (supported: txt, sqlite)", format)
			}

			return exporter.ExportChunks(input, writer)
		},
	}

	exportCmd.Flags().StringVarP(&format, "format", "f", "txt", "Output format: txt or sqlite")
	rootCmd.AddCommand(exportCmd)
}

func TestExportCmd_TextFormat(t *testing.T) {
	resetRootCmd()
	
	// Create temporary input file
	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	inputJSON := `{
		"chunkedPrompt": {
			"chunks": [
				{"text": "Test 1", "isThought": false},
				{"text": "Test 2", "isThought": false}
			]
		}
	}`
	if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
		t.Fatal(err)
	}
	tmpInput.Close()

	tmpOutput, err := os.CreateTemp("", "output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	// Run command
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", tmpInput.Name(), outputPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify output
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Test 1\n---\nTest 2"
	if string(result) != expected {
		t.Errorf("Result = %q, want %q", string(result), expected)
	}
}

func TestExportCmd_SQLiteFormat(t *testing.T) {
	resetRootCmd()
	
	// Create temporary input file
	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	inputJSON := `{
		"chunkedPrompt": {
			"chunks": [
				{"text": "SQLite Test 1", "isThought": false},
				{"text": "SQLite Test 2", "isThought": false}
			]
		}
	}`
	if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
		t.Fatal(err)
	}
	tmpInput.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Run command with sqlite format
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", tmpInput.Name(), dbPath, "-f", "sqlite"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify database content
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	var records []exporter.ChunkRecord
	if err := db.Find(&records).Error; err != nil {
		t.Fatalf("Failed to query records: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestExportCmd_UnsupportedFormat(t *testing.T) {
	resetRootCmd()
	
	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	inputJSON := `{"chunkedPrompt": {"chunks": []}}`
	if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
		t.Fatal(err)
	}
	tmpInput.Close()

	tmpOutput, err := os.CreateTemp("", "output_*")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	// Run command with invalid format
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", tmpInput.Name(), outputPath, "-f", "invalid"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
}

func TestExportCmd_InvalidInput(t *testing.T) {
	resetRootCmd()
	
	tmpOutput, err := os.CreateTemp("", "output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	// Run command with nonexistent input file
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", "/nonexistent/file.json", outputPath})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent input file, got nil")
	}
}

func TestExportCmd_MissingArgs(t *testing.T) {
	resetRootCmd()
	
	// Run command with missing arguments
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", "input.json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
}

func TestExportCmd_FormatAliases(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"text alias", "text"},
		{"db alias", "db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetRootCmd()
			
			tmpInput, err := os.CreateTemp("", "input_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpInput.Name())

			inputJSON := `{"chunkedPrompt": {"chunks": [{"text": "Test", "isThought": false}]}}`
			if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
				t.Fatal(err)
			}
			tmpInput.Close()

			var outputPath string
			if tt.format == "text" {
				tmpOutput, err := os.CreateTemp("", "output_*.txt")
				if err != nil {
					t.Fatal(err)
				}
				outputPath = tmpOutput.Name()
				tmpOutput.Close()
			} else {
				tmpDir := t.TempDir()
				outputPath = filepath.Join(tmpDir, "test.db")
			}
			defer os.Remove(outputPath)

			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs([]string{"export", tmpInput.Name(), outputPath, "-f", tt.format})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("Command failed for format %q: %v", tt.format, err)
			}
		})
	}
}

func TestExportCmd_WriteFailure(t *testing.T) {
	resetRootCmd()

	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	inputJSON := `{"chunkedPrompt": {"chunks": [{"text": "Test", "isThought": false}]}}`
	if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
		t.Fatal(err)
	}
	tmpInput.Close()

	// Try to write to a directory instead of a file (will cause error)
	tmpDir := t.TempDir()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", tmpInput.Name(), tmpDir})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when writing to directory, got nil")
	}
}

func TestExportCmd_DefaultFormat(t *testing.T) {
	resetRootCmd()

	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	inputJSON := `{"chunkedPrompt": {"chunks": [{"text": "Test", "isThought": false}]}}`
	if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
		t.Fatal(err)
	}
	tmpInput.Close()

	tmpOutput, err := os.CreateTemp("", "output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	// Don't specify format flag, should default to txt
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"export", tmpInput.Name(), outputPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify file exists and has content
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != "Test" {
		t.Errorf("Expected 'Test', got %q", string(result))
	}
}
