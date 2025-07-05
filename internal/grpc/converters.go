package grpc

import (
	"github.com/google/uuid"
	db "github.com/Lykeion-org/lexora-dictionary/internal/models"
	shared "github.com/Lykeion-org/go-shared/pkg/grpc/generated/shared"
)

// Proto to DB conversions
func convertProtoToReferent(r *shared.Referent) *db.Referent {
	if r == nil {
		return nil
	}
	
	uid, _ := uuid.Parse(r.GetUid())
	return &db.Referent{
		UID:         uid,
		EnReference: r.GetEnReference(),
		ImageSource: r.GetImageSource(),
	}
}

func convertProtoToSymbol(s *shared.Symbol) *db.Symbol {
	if s == nil {
		return nil
	}

	uid, _ := uuid.Parse(s.GetUid())
	var lemmaUID *uuid.UUID
	if s.GetLemma() != nil {
		uid, _ := uuid.Parse(s.GetLemma().GetUid())
		lemmaUID = &uid
	}

	return &db.Symbol{
		UID:        uid,
		Language:   int(s.GetLanguage()),
		SymbolType: int(s.GetSymbolType()),
		LemmaUID:   lemmaUID,
	}
}

func convertProtoToWord(w *shared.Word) *db.Word {
	if w == nil {
		return nil
	}

	uid, _ := uuid.Parse(w.GetUid())
	return &db.Word{
		UID:         uid,
		Word:        w.GetWord(),
		SoundSource: w.GetSoundSource(),
		IPA:         w.GetIpa(),
		WordType:    int(w.GetWordType()),
	}
}

// DB to Proto conversions
func convertReferentToProto(r *db.Referent) *shared.Referent {
	if r == nil {
		return nil
	}

	var symbols []*shared.Symbol
	for _, s := range r.Symbols {
		symbols = append(symbols, convertSymbolToProto(&s))
	}

	return &shared.Referent{
		Uid:         r.UID.String(),
		EnReference: r.EnReference,
		ImageSource: r.ImageSource,
		Symbols:     symbols,
	}
}

func convertSymbolToProto(s *db.Symbol) *shared.Symbol {
	if s == nil {
		return nil
	}

	return &shared.Symbol{
		Uid:        s.UID.String(),
		Language:   shared.Language(s.Language),
		SymbolType: shared.SymbolType(s.SymbolType),
		Lemma:      convertWordToProto(s.Lemma),
		Words:      convertWordListToProto(s.Words),
	}
}

func convertWordToProto(w *db.Word) *shared.Word {
	if w == nil {
		return nil
	}

	return &shared.Word{
		Uid:         w.UID.String(),
		Word:        w.Word,
		SoundSource: w.SoundSource,
		Ipa:         w.IPA,
		WordType:    shared.WordType(w.WordType),
	}
}

func convertWordListToProto(wl []db.Word) []*shared.Word {
	var words []*shared.Word
	for _, w := range wl {
		words = append(words, convertWordToProto(&w))
	}
	return words
}

func convertWordPointerListToProto(wl []*db.Word) []*shared.Word {
	var words []*shared.Word
	for _, w := range wl {
		words = append(words, convertWordToProto(w))
	}
	return words
}

func convertReferentListToProto(rl []db.Referent) []*shared.Referent {
	var referents []*shared.Referent
	for _, r := range rl {
		referents = append(referents, convertReferentToProto(&r))
	}
	return referents
}

// Helper functions
func derefString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
