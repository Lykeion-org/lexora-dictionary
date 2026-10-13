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
		Where("word ILIKE ?", "%"+word+"%").
		Find(&words).Error; err != nil {
		return nil, err
	}

	// Get unique word IDs
	wordIDs := make([]uuid.UUID, 0, len(words))
	for _, w := range words {
		wordIDs = append(wordIDs, w.UID)
	}
	if len(wordIDs) == 0 {
		return referents, nil
	}

	// Find referents through symbol-word relationships
	err := s.db.WithContext(ctx).
		Preload("Symbols.Lemma").
		Preload("Symbols.Words").
		Where(`referents.uid IN (
			SELECT referent_symbols.referent_id FROM referent_symbols
			JOIN symbol_words ON symbol_words.symbol_id = referent_symbols.symbol_id
			WHERE symbol_words.word_id IN ?)`, wordIDs).
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
