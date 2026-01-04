package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetDocumentByName retrieves a document by name.
func GetDocumentByName(ctx context.Context, db *gorm.DB, name string) (*Document, error) {
	var doc Document
	err := db.WithContext(ctx).Raw("SELECT * FROM documents WHERE name = ? LIMIT 1", name).Scan(&doc).Error
	if err != nil {
		return nil, err
	}
	if doc.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &doc, nil
}

// DeleteDocument deletes a document and all its associated chunks and embeddings.
func DeleteDocument(ctx context.Context, db *gorm.DB, docID string) error {
	// Delete embeddings for all chunks of the document
	if err := db.WithContext(ctx).Exec(`
		DELETE FROM chunk_embeddings 
		WHERE rowid IN (SELECT embedding_rowid FROM chunks WHERE document_id = ?)`, docID).Error; err != nil {
		return fmt.Errorf("failed to delete embeddings: %w", err)
	}

	// Delete chunks
	if err := db.WithContext(ctx).Where("document_id = ?", docID).Delete(&Chunk{}).Error; err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	// Delete document
	if err := db.WithContext(ctx).Delete(&Document{ID: docID}).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteDocumentByName deletes a document by name and all its associated chunks and embeddings.
func DeleteDocumentByName(ctx context.Context, db *gorm.DB, name string) error {
	doc, err := GetDocumentByName(ctx, db, name)
	if err != nil {
		return err
	}
	return DeleteDocument(ctx, db, doc.ID)
}

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
