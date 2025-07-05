package main

import (
	"log/slog"
	"github.com/Lykeion-org/go-shared/pkg/config"
	lyk_db "github.com/Lykeion-org/go-shared/pkg/db"
	"github.com/Lykeion-org/go-shared/pkg/helpers"
	"github.com/Lykeion-org/lexora-dictionary/internal/grpc"
	"github.com/Lykeion-org/lexora-dictionary/internal/models"
)



func main() {
	//Load config
	cfg, err := config.LoadConfigFile[Config]("../config.yaml")
	if err != nil {
		panic("Failed to load config file")
	}

	helpers.InitializeStandardLogger("debug")

	//Initialize database
	slog.Info("Attempting to connect to database")
	connectionString := lyk_db.CreateConnectionString(cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName, cfg.DatabaseUserName, cfg.DatabasePassword)
	db, err := lyk_db.ConnectAndMigrate(connectionString, &models.Referent{}, &models.ReferentSymbol{}, &models.Symbol{}, &models.SymbolWord{}, &models.Word{}, &models.WordAttributes{}, &models.WordSuggestion{}, &models.WordAttributesSuggestion{} )
	if err != nil {
		slog.Error("Cannot connect to database", "error", err)
		panic(err)
	}

	slog.Info("Connected to database")
	slog.Info("Opening grpc server")

	srv := grpc.NewLanguageServer(db)
	
	//Run main processes
	if(cfg.UseGrpcServer) {
		srv.StartServer(cfg.DictionaryServicePort); defer srv.StopServer()
	}

	select {}

}
