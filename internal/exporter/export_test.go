package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProcessChunks(t *testing.T) {
	tests := []struct {
		name     string
		root     Root
		expected string
	}{
		{
			name: "Regular chunks",
			root: Root{
				ChunkedPrompt: ChunkedPrompt{
					Chunks: []Chunk{
						{Text: "Hello", IsThought: false},
						{Text: "How are you?", IsThought: false},
					},
				},
			},
			expected: "Hello\n---\nHow are you?",
		},
		{
			name: "Ignore IsThought",
			root: Root{
				ChunkedPrompt: ChunkedPrompt{
					Chunks: []Chunk{
						{Text: "Thought", IsThought: true},
						{Text: "Answer", IsThought: false},
					},
				},
			},
			expected: "Answer",
		},
		{
			name: "Empty text",
			root: Root{
				ChunkedPrompt: ChunkedPrompt{
					Chunks: []Chunk{
						{Text: "", IsThought: false},
						{Text: "Text", IsThought: false},
					},
				},
			},
			expected: "Text",
		},
		{
			name: "All thoughts",
			root: Root{
				ChunkedPrompt: ChunkedPrompt{
					Chunks: []Chunk{
						{Text: "Thought 1", IsThought: true},
						{Text: "Thought 2", IsThought: true},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessChunks(tt.root)
			if got != tt.expected {
				t.Errorf("ProcessChunks() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExportChunks(t *testing.T) {
	// Create temporary input file
	inputJSON := `{
		"chunkedPrompt": {
			"chunks": [
				{"text": "Part 1", "isThought": false},
				{"text": "Thought", "isThought": true},
				{"text": "Part 2", "isThought": false}
			]
		}
	}`

	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	if _, err := tmpInput.Write([]byte(inputJSON)); err != nil {
		t.Fatal(err)
	}
	tmpInput.Close()

	// Create path for temporary output file
	tmpOutput, err := os.CreateTemp("", "output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	// Run export with TextWriter
	writer := &TextWriter{OutputPath: outputPath}
	err = ExportChunks(tmpInput.Name(), writer)
	if err != nil {
		t.Fatalf("ExportChunks failed: %v", err)
	}

	// Read result
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Part 1\n---\nPart 2"
	if string(result) != expected {
		t.Errorf("Result = %q, want %q", string(result), expected)
	}
}

func TestTextWriter(t *testing.T) {
	tmpOutput, err := os.CreateTemp("", "output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: []Chunk{
				{Text: "Line 1", IsThought: false},
				{Text: "Thought", IsThought: true},
				{Text: "Line 2", IsThought: false},
			},
		},
	}

	writer := &TextWriter{OutputPath: outputPath}
	if err := writer.Write(root); err != nil {
		t.Fatalf("TextWriter.Write failed: %v", err)
	}

	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Line 1\n---\nLine 2"
	if string(result) != expected {
		t.Errorf("Result = %q, want %q", string(result), expected)
	}
}

func TestTextWriter_WriteError(t *testing.T) {
	// Use an invalid path to trigger write error
	writer := &TextWriter{OutputPath: "/invalid/path/output.txt"}
	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: []Chunk{{Text: "Test", IsThought: false}},
		},
	}

	err := writer.Write(root)
	if err == nil {
		t.Error("Expected error when writing to invalid path, got nil")
	}
}

func TestSQLiteWriter(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: []Chunk{
				{Text: "Chunk 1", IsThought: false},
				{Text: "Thought", IsThought: true},
				{Text: "Chunk 2", IsThought: false},
				{Text: "", IsThought: false}, // Empty text should be filtered
			},
		},
	}

	writer := &SQLiteWriter{DBPath: dbPath}
	if err := writer.Write(root); err != nil {
		t.Fatalf("SQLiteWriter.Write failed: %v", err)
	}

	// Verify data in database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	var records []ChunkRecord
	if err := db.Find(&records).Error; err != nil {
		t.Fatalf("Failed to query records: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	expectedTexts := []string{"Chunk 1", "Chunk 2"}
	for i, record := range records {
		if record.Text != expectedTexts[i] {
			t.Errorf("Record %d: expected text %q, got %q", i, expectedTexts[i], record.Text)
		}
	}
}

func TestSQLiteWriter_InvalidPath(t *testing.T) {
	// Use an invalid path to trigger database error
	writer := &SQLiteWriter{DBPath: "/invalid/\x00/path/test.db"}
	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: []Chunk{{Text: "Test", IsThought: false}},
		},
	}

	err := writer.Write(root)
	if err == nil {
		t.Error("Expected error when opening invalid database path, got nil")
	}
}

func TestExportChunks_InvalidJSON(t *testing.T) {
	tmpInput, err := os.CreateTemp("", "input_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpInput.Name())

	// Write invalid JSON
	if _, err := tmpInput.Write([]byte("invalid json")); err != nil {
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

	writer := &TextWriter{OutputPath: outputPath}
	err = ExportChunks(tmpInput.Name(), writer)
	if err == nil {
		t.Error("Expected error when parsing invalid JSON, got nil")
	}
}

func TestExportChunks_FileNotFound(t *testing.T) {
	tmpOutput, err := os.CreateTemp("", "output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	outputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(outputPath)

	writer := &TextWriter{OutputPath: outputPath}
	err = ExportChunks("/nonexistent/file.json", writer)
	if err == nil {
		t.Error("Expected error when reading nonexistent file, got nil")
	}
}

func TestSQLiteWriter_EmptyChunks(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: []Chunk{},
		},
	}

	writer := &SQLiteWriter{DBPath: dbPath}
	if err := writer.Write(root); err != nil {
		t.Fatalf("SQLiteWriter.Write failed: %v", err)
	}

	// Verify no records in database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	var count int64
	if err := db.Model(&ChunkRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 records, got %d", count)
	}
}

func TestSQLiteWriter_AllThoughts(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: []Chunk{
				{Text: "Thought 1", IsThought: true},
				{Text: "Thought 2", IsThought: true},
			},
		},
	}

	writer := &SQLiteWriter{DBPath: dbPath}
	if err := writer.Write(root); err != nil {
		t.Fatalf("SQLiteWriter.Write failed: %v", err)
	}

	// Verify no records in database (all filtered out)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	var count int64
	if err := db.Model(&ChunkRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 records (all thoughts filtered), got %d", count)
	}
}

func TestSQLiteWriter_LargeDataset(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a large dataset
	chunks := make([]Chunk, 1000)
	for i := 0; i < 1000; i++ {
		chunks[i] = Chunk{
			Text:      fmt.Sprintf("Chunk %d", i),
			IsThought: i%2 == 0, // Half are thoughts
		}
	}

	root := Root{
		ChunkedPrompt: ChunkedPrompt{
			Chunks: chunks,
		},
	}

	writer := &SQLiteWriter{DBPath: dbPath}
	if err := writer.Write(root); err != nil {
		t.Fatalf("SQLiteWriter.Write failed: %v", err)
	}

	// Verify correct number of records
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	var count int64
	if err := db.Model(&ChunkRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	// Should have 500 records (half are thoughts, filtered out)
	if count != 500 {
		t.Errorf("Expected 500 records, got %d", count)
	}
}
