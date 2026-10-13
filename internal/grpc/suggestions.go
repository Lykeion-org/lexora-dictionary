package grpc

import (
	"context"
	"strings"

	lexora "github.com/Lykeion-org/go-shared/pkg/grpc/generated/lexora"
	shared "github.com/Lykeion-org/go-shared/pkg/grpc/generated/shared"
	db "github.com/Lykeion-org/lexora-dictionary/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultSuggestionPageSize = 50
const maxSuggestionPageSize = 500

// CreateWordSuggestion stores a suggestion. If the same word form is already
// pending for that language, the existing suggestion is returned instead.
func (s *LanguageServer) CreateWordSuggestion(ctx context.Context, req *lexora.CreateWordSuggestionRequest) (*lexora.WordSuggestion, error) {
	in := req.GetSuggestion()
	word := strings.TrimSpace(in.GetWord())
	if word == "" {
		return nil, status.Error(codes.InvalidArgument, "suggestion.word is required")
	}

	existing, err := s.SuggestionRepository.FindPendingSuggestion(ctx, word, int(in.GetLanguage()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}
	if existing != nil {
		return convertSuggestionToProto(existing), nil
	}

	suggestion := &db.WordSuggestion{
		UID:          uuid.New(),
		Word:         word,
		Lemma:        in.GetLemma(),
		Language:     int(in.GetLanguage()),
		SymbolType:   int(in.GetSymbolType()),
		WordType:     int(in.GetWordType()),
		JobUID:       in.GetJobUid(),
		RequesterUID: in.GetRequesterUid(),
	}
	if a := in.GetAttributes(); a != nil {
		suggestion.WordAttributes = db.WordAttributesSuggestion{
			Gender:         int(a.GetGender()),
			Number:         int(a.GetNumber()),
			Unique:         a.GetUnique(),
			Diminutive:     a.GetDiminutive(),
			Case:           int(a.GetCase()),
			Mood:           int(a.GetMood()),
			Tense:          int(a.GetTense()),
			Aspect:         int(a.GetAspect()),
			Person:         int(a.GetPerson()),
			DirectObject:   a.GetDirectObject(),
			IndirectObject: a.GetIndirectObject(),
			Valency:        int(a.GetValency()),
			Reflexive:      a.GetReflexive(),
		}
	}

	created, err := s.SuggestionRepository.CreateSuggestion(ctx, suggestion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}
	return convertSuggestionToProto(created), nil
}

// ListAllWordSuggestion lists pending (unapproved) suggestions, oldest first. page is 1-based.
func (s *LanguageServer) ListAllWordSuggestion(ctx context.Context, req *lexora.ListAllWordSuggestionRequest) (*lexora.ListAllWordSuggestionResponse, error) {
	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = defaultSuggestionPageSize
	}
	if pageSize > maxSuggestionPageSize {
		pageSize = maxSuggestionPageSize
	}

	suggestions, total, err := s.SuggestionRepository.ListPendingSuggestions(ctx, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	out := make([]*lexora.WordSuggestion, 0, len(suggestions))
	for i := range suggestions {
		out = append(out, convertSuggestionToProto(&suggestions[i]))
	}
	return &lexora.ListAllWordSuggestionResponse{
		Suggestions: out,
		TotalCount:  int32(total),
	}, nil
}

func convertSuggestionToProto(w *db.WordSuggestion) *lexora.WordSuggestion {
	a := w.WordAttributes
	return &lexora.WordSuggestion{
		Uid:          w.UID.String(),
		Word:         w.Word,
		Lemma:        w.Lemma,
		Language:     shared.Language(w.Language),
		SymbolType:   shared.SymbolType(w.SymbolType),
		WordType:     shared.WordType(w.WordType),
		JobUid:       w.JobUID,
		RequesterUid: w.RequesterUID,
		Approved:     w.Approved,
		ApprovedBy:   w.ApprovedBy,
		Attributes: &shared.WordAttributes{
			// The columns default to -1 ("not set"), which has no enum value; report it as unspecified.
			Gender:         shared.WordGender(nonNegative(a.Gender)),
			Number:         shared.WordNumber(nonNegative(a.Number)),
			Unique:         a.Unique,
			Diminutive:     a.Diminutive,
			Case:           shared.WordCase(nonNegative(a.Case)),
			Mood:           shared.WordMood(nonNegative(a.Mood)),
			Tense:          shared.WordTense(nonNegative(a.Tense)),
			Aspect:         shared.WordAspect(nonNegative(a.Aspect)),
			Person:         shared.WordPerson(nonNegative(a.Person)),
			DirectObject:   a.DirectObject,
			IndirectObject: a.IndirectObject,
			Valency:        int32(nonNegative(a.Valency)),
			Reflexive:      a.Reflexive,
		},
	}
}

func nonNegative(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
