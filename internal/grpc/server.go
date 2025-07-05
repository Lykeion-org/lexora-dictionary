package grpc

import (
	"net"
	"context"
	"errors"
	db "github.com/Lykeion-org/lexora-dictionary/internal/models"
	repository "github.com/Lykeion-org/lexora-dictionary/internal/repository"
	lexora "github.com/Lykeion-org/go-shared/pkg/grpc/generated/lexora"
	shared "github.com/Lykeion-org/go-shared/pkg/grpc/generated/shared"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type LanguageServer struct {
	lexora.UnimplementedDictionaryServiceServer
	ReferentRepository 			repository.ReferentRepository
	SymbolRepository   			repository.SymbolRepository
	WordRepository     			repository.WordRepository
	RelationshipRepository      repository.RelationshipRepository
	server						*grpc.Server
}

func NewLanguageServer(db *gorm.DB) *LanguageServer{
	return &LanguageServer{
		ReferentRepository: repository.NewReferentRepository(db),
		SymbolRepository: repository.NewSymbolRepository(db),
		WordRepository:      repository.NewWordRepository(db),
		RelationshipRepository: repository.NewRelationshipRepository(db),
	}
}

func (s *LanguageServer) StartServer(target string) error{
	grpcServer := grpc.NewServer()
	lexora.RegisterDictionaryServiceServer(grpcServer, s)

	listener, err := net.Listen("tcp", target)
	if err != nil {
		return err
	}

	go func(){
		if err := grpcServer.Serve(listener); err != nil{
			panic(err)
		}
	}()

	s.server = grpcServer
	return nil
}

func (s *LanguageServer) StopServer() error {
	s.server.GracefulStop()
	return nil
}


// READ
func (s *LanguageServer) GetReferent(ctx context.Context, req *lexora.GetReferentRequest) (*shared.Referent, error){
	//convert string to uid
	uid, err := uuid.Parse(req.GetUid()); if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	ref, err := s.ReferentRepository.GetReferent(ctx, uid); if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "referent not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertReferentToProto(ref), nil
}
func (s *LanguageServer) GetSymbol(ctx context.Context, req *lexora.GetSymbolRequest) (*shared.Symbol, error){
	uid, err := uuid.Parse(req.GetUid()); if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	sym, err := s.SymbolRepository.GetSymbol(ctx, uid); if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "symbol not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertSymbolToProto(sym), nil

}
func (s *LanguageServer) GetWord(ctx context.Context, req *lexora.GetWordRequest) (*shared.Word, error){
	uid, err := uuid.Parse(req.GetUid()); if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	wrd, err := s.WordRepository.GetWord(ctx, uid); if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "word not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertWordToProto(wrd), nil
}
func (s *LanguageServer) FindReferents(ctx context.Context, req *lexora.FindReferentsRequest) (*lexora.FindReferentsResponse, error){
	switch req.GetSearchMode(){
	case lexora.SearchMode_SEARCH_MODE_WORD:
		referents, err := s.ReferentRepository.ListReferentByWord(ctx, req.GetQuery() ); if err != nil {
			return nil, status.Errorf(codes.Internal, "db error: %v", err)
		}
		refList := convertReferentListToProto(referents)

		return &lexora.FindReferentsResponse{
			Referents: refList,
			TotalCount: int32(len(refList)),
		}, nil
	default:
		return nil, status.Error(codes.Unimplemented, "search mode not implemented")
	}
}

// CREATE
func (s *LanguageServer) CreateReferent(ctx context.Context, req *lexora.CreateReferentRequest) (*shared.Referent, error){
	var ref *db.Referent = &db.Referent{
		UID: uuid.New(),
		EnReference: req.EnReference,
		ImageSource: req.GetImageSource(),
	}
	
	createdRef, err := s.ReferentRepository.CreateReferent(ctx, ref); if err != nil {
		return nil, status.Errorf(codes.Internal, "db: error: %v", err)
	}

	return convertReferentToProto(createdRef), nil
}

func (s *LanguageServer) CreateSymbol(ctx context.Context, req *lexora.CreateSymbolRequest) (*shared.Symbol, error) {

	var sym *db.Symbol = &db.Symbol{
		UID: uuid.New(),
		Language: int(req.GetLanguage()),
		SymbolType: int(req.GetSymbolType()),
	}
	
	createdSym, err := s.SymbolRepository.CreateSymbol(ctx, sym); if err != nil {
		return nil, status.Errorf(codes.Internal, "db: error: %v", err)
	}

	return convertSymbolToProto(createdSym), nil
}

func (s *LanguageServer) CreateWord(ctx context.Context, req *lexora.CreateWordRequest) (*shared.Word, error) {

	var wrd *db.Word = &db.Word{
		UID: uuid.New(),
		Word: req.GetWord(),
		WordType: int(req.GetWordType()),
		SoundSource: req.GetSoundSource(),
		IPA: req.GetIpa(),
	}
	
	createdWrd, err := s.WordRepository.CreateWord(ctx, wrd); if err != nil {
		return nil, status.Errorf(codes.Internal, "db: error: %v", err)
	}

	return convertWordToProto(createdWrd), nil

}

// UPDATE
func (s *LanguageServer) UpdateReferent(ctx context.Context, req *lexora.UpdateReferentRequest) (*shared.Referent, error) {
	if req.GetReferent() == nil {
		return nil, status.Error(codes.InvalidArgument, "referent is required")
	}

	// First get existing referent
	uid, err := uuid.Parse(req.GetReferent().GetUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	existing, err := s.ReferentRepository.GetReferent(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "referent not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	// Update fields
	existing.EnReference = req.GetReferent().GetEnReference()
	if req.GetReferent().GetImageSource() != "" {
		imgSrc := req.GetReferent().GetImageSource()
		existing.ImageSource = imgSrc
	}

	// Save changes
	if err := s.ReferentRepository.UpdateReferent(ctx, existing); err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertReferentToProto(existing), nil
}
func (s *LanguageServer) UpdateSymbol(ctx context.Context, req *lexora.UpdateSymbolRequest) (*shared.Symbol, error) {
	if req.GetSymbol() == nil {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	// First get existing symbol
	uid, err := uuid.Parse(req.GetSymbol().GetUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	existing, err := s.SymbolRepository.GetSymbol(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "symbol not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	// Update fields
	existing.Language = int(req.GetSymbol().GetLanguage())
	existing.SymbolType = int(req.GetSymbol().GetSymbolType())
	// Note: ReferentUID cannot be updated directly through Symbol proto

	// Save changes
	if err := s.SymbolRepository.UpdateSymbol(ctx, existing); err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertSymbolToProto(existing), nil
}
func (s *LanguageServer) UpdateWord(ctx context.Context, req *lexora.UpdateWordRequest) (*shared.Word, error) {
	if req.GetWord() == nil {
		return nil, status.Error(codes.InvalidArgument, "word is required")
	}

	// First get existing word
	uid, err := uuid.Parse(req.GetWord().GetUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	existing, err := s.WordRepository.GetWord(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "word not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	// Update fields
	existing.Word = req.GetWord().GetWord()
	existing.WordType = int(req.GetWord().GetWordType())
	existing.SoundSource = req.GetWord().GetSoundSource()
	existing.IPA = req.GetWord().GetIpa()

	// Save changes
	if err := s.WordRepository.UpdateWord(ctx, existing); err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertWordToProto(existing), nil
}

// DELETE
func (s *LanguageServer) DeleteReferent(ctx context.Context, req *lexora.DeleteReferentRequest) (*shared.Referent, error) {
	uid, err := uuid.Parse(req.GetReferentUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	// First get the referent to return it after deletion
	ref, err := s.ReferentRepository.GetReferent(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "referent not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	// Delete the referent
	if err := s.ReferentRepository.DeleteReferent(ctx, uid); err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertReferentToProto(ref), nil
}
func (s *LanguageServer) DeleteSymbol(ctx context.Context, req *lexora.DeleteSymbolRequest) (*shared.Symbol, error) {
	uid, err := uuid.Parse(req.GetSymbolUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	// First get the symbol to return it after deletion
	sym, err := s.SymbolRepository.GetSymbol(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "symbol not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	// Delete the symbol
	if err := s.SymbolRepository.DeleteSymbol(ctx, uid); err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertSymbolToProto(sym), nil
}
func (s *LanguageServer) DeleteWord(ctx context.Context, req *lexora.DeleteWordRequest) (*shared.Word, error) {
	uid, err := uuid.Parse(req.GetWordUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UID: %v", err)
	}

	// First get the word to return it after deletion
	wrd, err := s.WordRepository.GetWord(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "word not found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	// Delete the word
	if err := s.WordRepository.DeleteWord(ctx, uid); err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return convertWordToProto(wrd), nil
}

// LINKING OPERATIONS
func (s *LanguageServer) LinkSymbolToReferent(ctx context.Context, req *lexora.LinkSymbolToReferentRequest) (*lexora.LinkSymbolToReferentResponse, error) {
	referentUID, err := uuid.Parse(req.GetReferentUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid referent UID: %v", err)
	}

	for _, symbolUIDStr := range req.GetSymbolUid() {
		symbolUID, err := uuid.Parse(symbolUIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid symbol UID: %v", err)
		}

		if err := s.RelationshipRepository.AddSymbolToReferent(ctx, referentUID, symbolUID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to link symbol to referent: %v", err)
		}
	}

	return &lexora.LinkSymbolToReferentResponse{Succes: true}, nil
}

func (s *LanguageServer) LinkWordToSymbol(ctx context.Context, req *lexora.LinkWordToSymbolRequest) (*lexora.LinkWordToSymbolResponse, error) {
	symbolUID, err := uuid.Parse(req.GetSymbolUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid symbol UID: %v", err)
	}

	for _, wordUIDStr := range req.GetWordUid() {
		wordUID, err := uuid.Parse(wordUIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid symbol UID: %v", err)
		}

		if err := s.RelationshipRepository.AddWordToSymbol(ctx, symbolUID, wordUID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to link symbol to referent: %v", err)
		}
	}

	return &lexora.LinkWordToSymbolResponse{Succes: true}, nil
}

func (s *LanguageServer) SetSymbolLemma(ctx context.Context, req *lexora.SetSymbolLemmaRequest) (*lexora.SetSymbolLemmaResponse, error) {
	wordUID, err := uuid.Parse(req.GetWordUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid word UID: %v", err)
	}

	symbolUID, err := uuid.Parse(req.GetSymbolUid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid symbol UID: %v", err)
	}

	if err := s.RelationshipRepository.SetSymbolLemma(ctx, symbolUID, wordUID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set symbol lemma: %v", err)
	}

	return &lexora.SetSymbolLemmaResponse{Succes: true}, nil
}
