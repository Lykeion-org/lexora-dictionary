package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/Lykeion-org/lexora-dictionary/internal/models"
)

type WordAttributesRepository interface {
	CreateWordAttributes(ctx context.Context, attr *models.WordAttributes) (*models.WordAttributes, error)
	GetWordAttributes(ctx context.Context, wordUID uuid.UUID) (*models.WordAttributes, error)
	UpdateWordAttributes(ctx context.Context, attr *models.WordAttributes) error
	DeleteWordAttributes(ctx context.Context, wordUID uuid.UUID) error
}

type wordAttributesRepository struct {
	db *gorm.DB
}

func NewWordAttributesRepository(db *gorm.DB) *wordAttributesRepository {
	return &wordAttributesRepository{
		db: db,
	}
}

func (r *wordAttributesRepository) CreateWordAttributes(ctx context.Context, attr *models.WordAttributes) (*models.WordAttributes, error) {
	if err := r.db.WithContext(ctx).Create(attr).Error; err != nil {
		return nil, err
	}
	return attr, nil
}

func (r *wordAttributesRepository) UpdateWordAttributes(ctx context.Context, attr *models.WordAttributes) error {
	return r.db.WithContext(ctx).Save(attr).Error
}

func (r *wordAttributesRepository) DeleteWordAttributes(ctx context.Context, wordUID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.WordAttributes{}, "word_id = ?", wordUID).Error
}

func (r *wordAttributesRepository) GetWordAttributes(ctx context.Context, wordUID uuid.UUID) (*models.WordAttributes, error) {
	var attr models.WordAttributes
	if err := r.db.WithContext(ctx).First(&attr, "word_id = ?", wordUID).Error; err != nil {
		return nil, err
	}
	return &attr, nil
}
