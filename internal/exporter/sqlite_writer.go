package exporter

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ChunkRecord represents a chunk record in the database.
type ChunkRecord struct {
	ID   uint   `gorm:"primaryKey"`
	Text string `gorm:"not null"`
}

// SQLiteWriter writes chunks to a SQLite database using GORM.
type SQLiteWriter struct {
	DBPath string
}

// Write writes the chunks to a SQLite database.
func (w *SQLiteWriter) Write(root Root) error {
	db, err := gorm.Open(sqlite.Open(w.DBPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err := db.AutoMigrate(&ChunkRecord{}); err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}

	records := make([]ChunkRecord, 0, len(root.ChunkedPrompt.Chunks))
	for _, chunk := range root.ChunkedPrompt.Chunks {
		if chunk.IsThought {
			continue
		}
		if chunk.Text != "" {
			records = append(records, ChunkRecord{
				Text: chunk.Text,
			})
		}
	}

	// Only insert if there are records
	if len(records) > 0 {
		if err := db.Create(&records).Error; err != nil {
			return fmt.Errorf("error inserting chunks: %w", err)
		}
	}

	return nil
}
