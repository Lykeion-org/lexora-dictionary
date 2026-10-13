package repository

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/Lykeion-org/lexora-dictionary/internal/models"
)

type SuggestionRepository interface {
	CreateSuggestion(ctx context.Context, suggestion *models.WordSuggestion) (*models.WordSuggestion, error)
	GetSuggestion(ctx context.Context, uid uuid.UUID) (*models.WordSuggestion, error)
	UpdateSuggestion(ctx context.Context, suggestion *models.WordSuggestion) error
	DeleteSuggestion(ctx context.Context, uid uuid.UUID) error
	ListPendingSuggestions(ctx context.Context, page int, pageSize int) ([]models.WordSuggestion, int64, error)
	FindPendingSuggestion(ctx context.Context, word string, language int) (*models.WordSuggestion, error)
	ApproveSuggestion(ctx context.Context, uid uuid.UUID, approvedBy string) error
}

type suggestionRepository struct {
	db *gorm.DB
}

func NewSuggestionRepository(db *gorm.DB) *suggestionRepository {
	return &suggestionRepository{
		db: db,
	}
}

func (r *suggestionRepository) CreateSuggestion(ctx context.Context, suggestion *models.WordSuggestion) (*models.WordSuggestion, error) {
	if suggestion.UID == uuid.Nil {
		suggestion.UID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(suggestion).Error; err != nil {
		return nil, err
	}
	return suggestion, nil
}

func (r *suggestionRepository) UpdateSuggestion(ctx context.Context, suggestion *models.WordSuggestion) error {
	return r.db.WithContext(ctx).Save(suggestion).Error
}

func (r *suggestionRepository) DeleteSuggestion(ctx context.Context, uid uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.WordSuggestion{}, "uid = ?", uid).Error
}

func (r *suggestionRepository) GetSuggestion(ctx context.Context, uid uuid.UUID) (*models.WordSuggestion, error) {
	var suggestion models.WordSuggestion
	if err := r.db.WithContext(ctx).Preload("WordAttributes").First(&suggestion, "uid = ?", uid).Error; err != nil {
		return nil, err
	}
	return &suggestion, nil
}

// ListPendingSuggestions returns one page (1-based) of unapproved suggestions, oldest first, and the total count.
func (r *suggestionRepository) ListPendingSuggestions(ctx context.Context, page int, pageSize int) ([]models.WordSuggestion, int64, error) {
	var suggestions []models.WordSuggestion
	var total int64

	query := r.db.WithContext(ctx).Model(&models.WordSuggestion{}).Where("approved = ?", false)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("WordAttributes").
		Order("created_at ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&suggestions).Error
	return suggestions, total, err
}

// FindPendingSuggestion returns the unapproved suggestion for this word form and language, or nil if there is none.
func (r *suggestionRepository) FindPendingSuggestion(ctx context.Context, word string, language int) (*models.WordSuggestion, error) {
	var suggestions []models.WordSuggestion
	err := r.db.WithContext(ctx).
		Preload("WordAttributes").
		Where("approved = ? AND language = ? AND LOWER(word) = LOWER(?)", false, language, word).
		Limit(1).
		Find(&suggestions).Error
	if err != nil || len(suggestions) == 0 {
		return nil, err
	}
	return &suggestions[0], nil
}

func (r *suggestionRepository) ApproveSuggestion(ctx context.Context, uid uuid.UUID, approvedBy string) error {
	return r.db.WithContext(ctx).
		Model(&models.WordSuggestion{}).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"approved": true,
			"approved_by": approvedBy,
		}).Error
}
