package db

import (
	"time"
	"github.com/google/uuid"
)

type Referent struct {
	UID         		uuid.UUID 					`gorm:"type:uuid;index;default:uuid_generate_v4();primaryKey"`
	EnReference 		string    					`gorm:"not null"`
	ImageSource 		string			
	CreatedAt   		time.Time
	UpdatedAt 			time.Time

	//Many-to-many relations
	Symbols     		[]Symbol 					`gorm:"many2many:referent_symbols"`
}

type Symbol struct {
	UID       	 		uuid.UUID 					`gorm:"type:uuid;index;default:uuid_generate_v4();primaryKey"`
	Language   			int       					`gorm:"not null"`
	SymbolType 			int       					`gorm:"not null"`
	LemmaUID   			*uuid.UUID
	Lemma      			*Word  						`gorm:"foreignKey:LemmaUID"`
	Words      			[]Word 						`gorm:"many2many:symbol_words"`
	CreatedAt   		time.Time
	UpdatedAt 			time.Time
}

type ReferentSymbol struct {
	ReferentID 			uuid.UUID 					`gorm:"type:uuid;primaryKey;column:referent_id"`
	SymbolID 			uuid.UUID 					`gorm:"type:uuid;primaryKey;column:symbol_id"`
	CreatedAt   		time.Time
	UpdatedAt 			time.Time
}


type Word struct {
	UID         		uuid.UUID	 				`gorm:"type:uuid;index;default:uuid_generate_v4();primaryKey"`
	Word        		string   	 				`gorm:"not null"`
	SoundSource 		string
	IPA         		string
	WordType    		int
	CreatedAt   		time.Time
	UpdatedAt 			time.Time

	//One-to-one relationship
	WordAttributes 		WordAttributes 				`gorm:"foreignKey:WordID"`
}

type SymbolWord struct{
	SymbolID			uuid.UUID					`gorm:"type:uuid;primaryKey"`
	WordID 				uuid.UUID					`gorm:"type:uuid;primaryKey"`
	CreatedAt   		time.Time
	UpdatedAt 			time.Time
}

type WordAttributes struct {
	WordID   			uuid.UUID 					`gorm:"type:uuid;index;primaryKey"`
	Gender   			int							`gorm:"default:-1"`
	Number 				int							`gorm:"default:-1"`
	Unique 				bool						`gorm:"default:false"`
	Diminutive 			bool						`gorm:"default:false"`
	Case 				int							`gorm:"default:-1"`
	Mood 				int							`gorm:"default:-1"`
	Tense 				int							`gorm:"default:-1"`
	Aspect 				int							`gorm:"default:-1"`
	Person 				int							`gorm:"default:-1"`
	IndirectObject 		bool						`gorm:"default:false"`
	Valency 			int							`gorm:"default:-1"`
	Reflexive 			bool						`gorm:"default:false"`
	CreatedAt   		time.Time
	UpdatedAt 			time.Time
}

type WordSuggestion struct {
	UID         		uuid.UUID 					`gorm:"type:uuid;primaryKey"`
	Word        		string    					`gorm:"not null"`
	SoundSource 		string
	IPA         		string
	WordType    		int
	Approved			bool						`gorm:"default:false"`
	ApprovedBy			string
	CreatedAt   		time.Time
	UpdatedAt 			time.Time
	
	//One-to-one relationship
	WordAttributes 		WordAttributesSuggestion 	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type WordAttributesSuggestion struct {
	WordSuggestionID   	uuid.UUID 					`gorm:"type:uuid;primaryKey"`
	Gender   			int							`gorm:"default:-1"`
	Number 				int							`gorm:"default:-1"`
	Unique 				bool						`gorm:"default:-1"`	
	Diminutive 			bool						`gorm:"default:false"`	
	Case 				int							`gorm:"default:false"`
	Mood 				int							`gorm:"default:-1"`
	Tense 				int							`gorm:"default:-1"`
	Aspect 				int							`gorm:"default:-1"`
	Person 				int							`gorm:"default:-1"`
	DirectObject 		bool						`gorm:"default:false"`	
	IndirectObject 		bool						`gorm:"default:false"`	
	Valency 			int							`gorm:"default:-1"`
	Reflexive 			bool						`gorm:"default:false"`
	Approved			bool						`gorm:"default:false"`
	ApprovedBy			string
	CreatedAt   		time.Time
	UpdatedAt 			time.Time
}
