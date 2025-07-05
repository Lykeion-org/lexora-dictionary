package repository

import (
	"context"

	"github.com/Lykeion-org/lexora-dictionary/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReferentRepository interface {
	CreateReferent(ctx context.Context, ref *models.Referent) (*models.Referent, error)
	GetReferent(ctx context.Context, uid uuid.UUID) (*models.Referent, error)
	UpdateReferent(ctx context.Context, ref *models.Referent) error
	DeleteReferent(ctx context.Context, uid uuid.UUID) error
	ListReferentByWord(ctx context.Context, word string) ([]models.Referent, error)
	ListReferents(ctx context.Context, limit, offset int) ([]models.Referent, error)
	FindReferents(ctx context.Context, searchTerm string) ([]models.Referent, error)
}
type referentRepository struct {
	db *gorm.DB
}

func NewReferentRepository(db *gorm.DB) *referentRepository {
	return &referentRepository{
		db: db,
	}
}

func (s *referentRepository) CreateReferent(ctx context.Context, ref *models.Referent) (*models.Referent, error) {
	if ref.UID == uuid.Nil {
		ref.UID = uuid.New()
	}
	if err := s.db.WithContext(ctx).Create(ref).Error; err != nil {
		return nil, err
	}
	return ref, nil
}
func (s *referentRepository) UpdateReferent(ctx context.Context, ref *models.Referent) error {
	return s.db.WithContext(ctx).Save(ref).Error
}

func (s *referentRepository) DeleteReferent(ctx context.Context, uid uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&models.Referent{}, "uid = ?", uid).Error
}

func (s *referentRepository) GetReferent(ctx context.Context, uid uuid.UUID) (*models.Referent, error) {
	var data models.Referent
	if err := s.db.
		WithContext(ctx).
		Preload("Symbols.Lemma").
		Preload("Symbols.Words").
		Take(&data, "uid = ?", uid).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *referentRepository) ListReferentByWord(ctx context.Context, word string) ([]models.Referent, error) {
	var referents []models.Referent

	// First find matching words
	var words []models.Word
	if err := s.db.WithContext(ctx).
		Where("word LIKE ?", "%"+word+"%").
		Find(&words).Error; err != nil {
		return nil, err
	}

	// Get unique word IDs
	wordIDs := make([]uuid.UUID, 0, len(words))
	for _, w := range words {
		wordIDs = append(wordIDs, w.UID)
	}

	// Find referents through symbol-word relationships
	err := s.db.WithContext(ctx).
		Preload("Symbols.Words").
		Joins("JOIN referent_symbols ON referents.uid = referent_symbols.referent_uid").
		Joins("JOIN symbols ON referent_symbols.symbol_uid = symbols.uid").
		Joins("JOIN symbol_words ON symbols.uid = symbol_words.symbol_uid").
		Where("symbol_words.word_uid IN ?", wordIDs).
		Distinct("referents.uid").
		Find(&referents).Error

	return referents, err
}

func (s *referentRepository) ListReferents(ctx context.Context, limit, offset int) ([]models.Referent, error) {
	var referents []models.Referent

	query := s.db.WithContext(ctx).Preload("Symbols.Words")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&referents).Error
	return referents, err
}

func (s *referentRepository) FindReferents(ctx context.Context, searchTerm string) ([]models.Referent, error) {
	var referents []models.Referent

	// Search in EnReference field and ImageSource field
	err := s.db.WithContext(ctx).
		Preload("Symbols.Words").
		Where("en_reference ILIKE ? OR image_source ILIKE ?", "%"+searchTerm+"%", "%"+searchTerm+"%").
		Find(&referents).Error

	return referents, err
}
