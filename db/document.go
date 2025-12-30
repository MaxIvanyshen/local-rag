package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SaveDocument creates a new document in the database and returns its ID.
func SaveDocument(ctx context.Context, db *gorm.DB, name string) (string, error) {
	docID := uuid.New().String()

	doc := Document{
		ID:   docID,
		Name: name,
	}

	if err := db.WithContext(ctx).Create(&doc).Error; err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return docID, nil
}
