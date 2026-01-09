package exporter

import (
	"os"
	"testing"
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

	// Run export
	err = ExportChunks(tmpInput.Name(), outputPath)
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
