package repository

import (
	"context"
    "github.com/google/uuid"
    "gorm.io/gorm"
	"github.com/Lykeion-org/lexora-dictionary/internal/models"
)
type WordRepository interface {
	CreateWord(ctx context.Context, wrd *models.Word) (*models.Word, error)
	GetWord(ctx context.Context, uid uuid.UUID) (*models.Word, error)
	GetWordAttributes(ctx context.Context, wordID uuid.UUID) (*models.WordAttributes, error)
	UpdateWord(ctx context.Context, wrd *models.Word) error
	DeleteWord(ctx context.Context, uid uuid.UUID) error
	ListWordByWordAttribute(ctx context.Context, sourceWordUID uuid.UUID, targetLanguage int) ([]*models.Word, error)
	CreateWordAttributes(ctx context.Context, attrs *models.WordAttributes) error
}
type wordRepository struct {
	db *gorm.DB
}

func NewWordRepository(db *gorm.DB) *wordRepository{
	return &wordRepository{
		db: db,
	}
}

func (s *wordRepository) CreateWord(ctx context.Context, wrd *models.Word) (*models.Word, error) {
	if wrd.UID == uuid.Nil {
        wrd.UID = uuid.New()
    }
    if err := s.db.WithContext(ctx).Create(wrd).Error; err != nil {
        return nil, err
    }
    return wrd, nil
}
func (s *wordRepository) UpdateWord(ctx context.Context, wrd *models.Word) error {
	if err := s.db.WithContext(ctx).Save(wrd).Error; err != nil {
		return err
	}
	
	// Update associated WordAttributes if they exist
	if wrd.WordAttributes.WordID != uuid.Nil {
		return s.db.WithContext(ctx).Model(&models.WordAttributes{}).
			Where("word_id = ?", wrd.UID).
			Updates(wrd.WordAttributes).Error
	}
	return nil
}

func (s *wordRepository) DeleteWord(ctx context.Context, uid uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&models.Word{}, "uid = ?", uid).Error
}

func (s *wordRepository) GetWord(ctx context.Context, uid uuid.UUID) (*models.Word, error) {
	var data models.Word
    if err := s.db.WithContext(ctx).First(&data, "uid = ?", uid).Error; err != nil {
        return nil, err
    }
    return &data, nil
}

func (s *wordRepository) GetWordAttributes(ctx context.Context, wordID uuid.UUID) (*models.WordAttributes, error) {
	var attrs models.WordAttributes
	if err := s.db.WithContext(ctx).First(&attrs, "word_id = ?", wordID).Error; err != nil {
		return nil, err
	}
	return &attrs, nil
}

// ListWordByWordAttribute finds words in the target language that have the same attributes as the source word.
// It follows the relationship chain: Word -> Symbol -> Referent -> Symbol (target) -> Word (target)
// and matches the word attributes between source and target words.
func (s *wordRepository) CreateWordAttributes(ctx context.Context, attrs *models.WordAttributes) error {
	if attrs.WordID == uuid.Nil {
		return gorm.ErrInvalidData
	}
	return s.db.WithContext(ctx).Create(attrs).Error
}

func (s *wordRepository) ListWordByWordAttribute(ctx context.Context, sourceWordUID uuid.UUID, targetLanguage int) ([]*models.Word, error) {
    
    // First get the source word with its attributes
    var sourceWord models.Word
    if err := s.db.WithContext(ctx).
        Preload("WordAttributes").
        First(&sourceWord, "uid = ?", sourceWordUID).Error; err != nil {
        return nil, err
    }

    // Find all symbols containing the source word
    var symbols []models.Symbol
    if err := s.db.WithContext(ctx).
        Joins("JOIN symbol_words sw ON sw.symbol_id = symbols.uid").
        Where("sw.word_id = ?", sourceWordUID).
        Find(&symbols).Error; err != nil {
        return nil, err
    }

    // For each symbol, find its referent and then find other symbols in the target language
    var targetWords []*models.Word
    for _, symbol := range symbols {
        var referent models.Referent
        if err := s.db.WithContext(ctx).
            Joins("JOIN referent_symbols rs ON rs.referent_id = referents.uid").
            Where("rs.symbol_id = ?", symbol.UID).
            First(&referent).Error; err != nil {
            continue // Skip if we can't find the referent
        }

        // Find symbols in target language for this referent
        var targetSymbols []models.Symbol
        if err := s.db.WithContext(ctx).
            Joins("JOIN referent_symbols rs ON rs.symbol_id = symbols.uid").
            Where("rs.referent_id = ? AND symbols.language = ?", referent.UID, targetLanguage).
            Find(&targetSymbols).Error; err != nil {
            continue
        }

        // For each target symbol, find words with matching attributes
        for _, targetSymbol := range targetSymbols {
            var words []*models.Word
            if err := s.db.WithContext(ctx).
                Preload("WordAttributes").
                Joins("JOIN symbol_words sw ON sw.word_id = words.uid").
                Where("sw.symbol_id = ?", targetSymbol.UID).
                Where(`EXISTS (
                    SELECT 1 FROM word_attributes wa 
                    WHERE wa.word_id = words.uid 
                    AND wa.gender = ? 
                    AND wa.number = ? 
                    AND wa.unique = ? 
                    AND wa.diminutive = ? 
                    AND wa.case = ? 
                    AND wa.mood = ? 
                    AND wa.tense = ? 
                    AND wa.aspect = ? 
                    AND wa.person = ? 
                    AND wa.indirect_object = ? 
                    AND wa.valency = ? 
                    AND wa.reflexive = ?
                )`, 
                sourceWord.WordAttributes.Gender,
                sourceWord.WordAttributes.Number,
                sourceWord.WordAttributes.Unique,
                sourceWord.WordAttributes.Diminutive,
                sourceWord.WordAttributes.Case,
                sourceWord.WordAttributes.Mood,
                sourceWord.WordAttributes.Tense,
                sourceWord.WordAttributes.Aspect,
                sourceWord.WordAttributes.Person,
                sourceWord.WordAttributes.IndirectObject,
                sourceWord.WordAttributes.Valency,
                sourceWord.WordAttributes.Reflexive).
                Find(&words).Error; err != nil {
                continue
            }
            targetWords = append(targetWords, words...)
        }
    }

    return targetWords, nil
}
