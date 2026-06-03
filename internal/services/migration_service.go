package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/primesoftwaresi/prime-migration-fyne/internal/models"
	"github.com/primesoftwaresi/prime-migration/pkg/db"
	"github.com/primesoftwaresi/prime-migration/pkg/db/historydb"
	"github.com/primesoftwaresi/prime-migration/pkg/db/prisma5"
	"github.com/primesoftwaresi/prime-migration-fyne/pkg/logger"
)

type MigrationService struct {
	history *HistoryService
}

func NewMigrationService(history *HistoryService) *MigrationService {
	return &MigrationService{history: history}
}

func (ms *MigrationService) ConnectDatabase(config models.DatabaseConfig) (*sql.DB, error) {
	dbConfig := db.DbConfig{
		Host:      config.Host,
		Port:      config.Port,
		Path:      config.Path,
		User:      config.User,
		Password:  config.Password,
		Conversao: config.Conversao,
	}
	logger.Info("Conectando ao banco de dados: Host=%s, Port=%s, Path=%s", config.Host, config.Port, config.Path)
	database, err := db.Connect(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %w", err)
	}
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("erro ao testar conexão: %w", err)
	}
	logger.Info("Conexão estabelecida com sucesso: %s", config.Path)
	return database, nil
}

func (ms *MigrationService) SetupProcedures(database *sql.DB, sistema, clienteCodigo string) error {
	logger.Info("Configurando procedures para %s, cliente %s", sistema, clienteCodigo)
	
	// PRISMA5 agora usa código Go direto, não precisa de procedures SQL
	if strings.ToUpper(sistema) == "PRISMA5" {
		return prisma5.SetupPRISMA5Database(database)
	}
	
	// FCERTA ainda usa procedures SQL
	return db.SetupProcedures(database, sistema, clienteCodigo)
}

func (ms *MigrationService) SetupPharmacieConnection(database *sql.DB, config *models.MigrationConfig) error {
	if config.System != "FCERTA" && config.System != "PRISMA5" {
		return nil
	}
	_, err := database.Exec(
		"INSERT INTO CONEXAO (ALIAS, IPSERVER, PORTA) VALUES (?, ?, ?)",
		config.AliasPharmacie, config.IpServerPharmacie, config.PortaPharmacie,
	)
	if err != nil {
		if !containsError(err, "unique") && !containsError(err, "duplicate") {
			return fmt.Errorf("erro ao inserir dados Pharmacie: %w", err)
		}
	}
	return nil
}

func (ms *MigrationService) ImportExcel(database *sql.DB, excelPath string) error {
	if excelPath == "" {
		return nil
	}
	return prisma5.ImportarPlanilhaGruposPRISMA5(database, excelPath)
}

func (ms *MigrationService) MigrateOption(
	ctx context.Context,
	config *models.MigrationConfig,
	option string,
	runCtx *MigrationRunContext,
	onProgress func(message string),
) (*prisma5.OptionStats, error) {
	onProgress(fmt.Sprintf("Iniciando migração de %s...", ms.GetOptionName(option, config.System)))
	dbConfig := db.DbConfig{
		Host:      config.Database.Host,
		Port:      config.Database.Port,
		Path:      config.Database.Path,
		User:      config.Database.User,
		Password:  config.Database.Password,
		Conversao: config.Database.Conversao,
	}
	database, err := db.Connect(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %w", err)
	}
	defer database.Close()

	incremental := config.Mode == models.MigrationModeIncremental
	extra := &db.MigrarExtras{Incremental: incremental}

	if ms.history != nil && ms.history.Connected() && runCtx != nil && config.System == "PRISMA5" {
		entity := historydb.EntityTypeFromOption(option)
		meta, err := ms.history.GetMigratedRecordMeta(runCtx.ClientID, entity)
		if err != nil {
			logger.Warn("histórico: meta incremental indisponível: %v", err)
			meta = nil
		}
		cb := &prisma5.MigrationCallbacks{
			Incremental: incremental,
			EntityType:  entity,
			RunID:       runCtx.RunID,
			ClientID:    runCtx.ClientID,
			RecordMeta:  meta,
			Stats:       &prisma5.OptionStats{},
			SaveRecord: func(runID string, clientID, entityType int, sourceID, destinationID, hash string) error {
				return ms.history.SaveRecord(runID, clientID, entityType, sourceID, destinationID, hash)
			},
			SaveError: func(runID string, entityType int, sourceID, errMsg string) error {
				return ms.history.SaveError(runID, entityType, sourceID, errMsg)
			},
		}
		extra.PRISMACallbacks = cb
	}

	if err := db.MigrarDados(database, option, config.Database.Conversao, config.VVencido, config.System, extra); err != nil {
		return extraStats(extra), fmt.Errorf("erro na migração: %w", err)
	}
	onProgress(fmt.Sprintf("%s migrado com sucesso!", ms.GetOptionName(option, config.System)))
	return extraStats(extra), nil
}

func extraStats(extra *db.MigrarExtras) *prisma5.OptionStats {
	if extra != nil && extra.PRISMACallbacks != nil && extra.PRISMACallbacks.Stats != nil {
		return extra.PRISMACallbacks.Stats
	}
	return nil
}

func (ms *MigrationService) GetOptionName(option, sistema string) string {
	switch option {
	case "1":
		return "Clientes"
	case "2":
		return "Médicos"
	case "3":
		return "Fornecedores"
	case "4":
		return "Produtos"
	case "5":
		if sistema == "FCERTA" {
			return "Lotes"
		}
		return "Forma Farmacêutica"
	case "6":
		if sistema == "FCERTA" {
			return "Produção Interna"
		}
		return "Lotes"
	case "7":
		if sistema == "FCERTA" {
			return "Refazer - Histórico Vendas"
		}
		return "Produção Interna"
	case "8":
		if sistema == "PRISMA5" {
			return "Refazer - Histórico Vendas"
		}
		return "Forma Farmacêutica"
	default:
		return option
	}
}

func containsError(err error, substr string) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	substr = strings.ToLower(substr)
	return strings.Contains(errStr, substr)
}
