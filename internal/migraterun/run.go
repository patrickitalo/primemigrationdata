// Package migraterun executa o pipeline completo de migração (histórico, conexão, setup, opções).
// Usado pela UI Fyne e pelo binário cmd/pm-migrate.
package migraterun

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/primesoftwaresi/prime-migration-fyne/internal/models"
	"github.com/primesoftwaresi/prime-migration-fyne/internal/services"
	"github.com/primesoftwaresi/prime-migration/pkg/db"
	"github.com/primesoftwaresi/prime-migration/pkg/db/historydb"
	"github.com/primesoftwaresi/prime-migration-fyne/pkg/logger"
)

// Params define a execução da migração.
type Params struct {
	Ctx             context.Context
	Config          *models.MigrationConfig
	UserSession     *models.UserSession
	Migration       *services.MigrationService
	History         *services.HistoryService
	OnLog           func(string)
	OnOptionStatus  func(index int, status models.MigrationStatus)
	OnProgress      func(done, total int)
}

// Result resume o desfecho (erros parciais não são erro fatal).
type Result struct {
	HasPartialErrors bool
}

// Run executa o mesmo fluxo que a janela de migração (conexão, setup, loop de opções).
func Run(p Params) (*Result, error) {
	res := &Result{}
	if p.OnLog == nil {
		p.OnLog = func(string) {}
	}
	if p.Ctx == nil {
		p.Ctx = context.Background()
	}
	cfg := p.Config
	if cfg == nil {
		return nil, fmt.Errorf("configuração ausente")
	}

	if err := db.InitConfigDB(); err != nil {
		p.OnLog(fmt.Sprintf("Erro ao inicializar SQLite: %v", err))
		return nil, err
	}
	defer db.CloseConfigDB()

	logger.SetLogSQLiteHook(func(level, message, codigoCliente, sistemaOrigem, tipoMigracao string) {
		_ = db.SaveLog(level, message, codigoCliente, sistemaOrigem, tipoMigracao)
	})
	logger.SetLogContext(cfg.ClientCode, cfg.System, "")

	p.OnLog(fmt.Sprintf("Iniciando migração para cliente %s", cfg.ClientCode))
	p.OnLog(fmt.Sprintf("Sistema: %s", cfg.System))

	if p.History == nil || !p.History.Connected() {
		p.OnLog("Histórico central não conectado. Migração abortada.")
		return nil, fmt.Errorf("histórico central não conectado")
	}

	p.OnLog("Conectando ao banco de dados de origem...")
	p.OnLog(fmt.Sprintf("   Host: %s, Porta: %s", cfg.Database.Host, cfg.Database.Port))
	p.OnLog(fmt.Sprintf("   Path: %s", cfg.Database.Path))

	database, err := p.Migration.ConnectDatabase(cfg.Database)
	if err != nil {
		p.OnLog(fmt.Sprintf("Erro ao conectar: %v", err))
		return nil, err
	}
	defer database.Close()
	p.OnLog(fmt.Sprintf("Conectado ao banco: %s", cfg.Database.Path))

	select {
	case <-p.Ctx.Done():
		p.OnLog("Migração cancelada pelo usuário")
		return res, p.Ctx.Err()
	default:
	}

	p.OnLog("Configurando procedures / DDL...")
	if err := p.Migration.SetupProcedures(database, cfg.System, cfg.ClientCode); err != nil {
		p.OnLog(fmt.Sprintf("Erro ao configurar procedures: %v", err))
		return nil, err
	}

	select {
	case <-p.Ctx.Done():
		p.OnLog("Migração cancelada pelo usuário")
		return res, p.Ctx.Err()
	default:
	}

	p.OnLog("Procedures / DDL configurados")

	select {
	case <-p.Ctx.Done():
		p.OnLog("Migração cancelada pelo usuário")
		return res, p.Ctx.Err()
	default:
	}

	if cfg.System == "FCERTA" || cfg.System == "PRISMA5" {
		err = p.Migration.SetupPharmacieConnection(database, cfg)
		if err != nil {
			p.OnLog(fmt.Sprintf("Aviso ao configurar Pharmacie: %v", err))
		} else {
			p.OnLog("Conexão Pharmacie configurada")
		}
	}

	select {
	case <-p.Ctx.Done():
		p.OnLog("Migração cancelada pelo usuário")
		return res, p.Ctx.Err()
	default:
	}

	if cfg.System == "PRISMA5" && cfg.ExcelPath != "" {
		p.OnLog("Importando planilha Excel...")
		err = p.Migration.ImportExcel(database, cfg.ExcelPath)
		if err != nil {
			p.OnLog(fmt.Sprintf("Erro ao importar Excel: %v", err))
		} else {
			p.OnLog("Planilha importada")
		}
	}

	select {
	case <-p.Ctx.Done():
		p.OnLog("Migração cancelada pelo usuário")
		return res, p.Ctx.Err()
	default:
	}

	implantador := ""
	if p.UserSession != nil {
		if p.UserSession.Nick != "" {
			implantador = p.UserSession.Nick
		} else {
			implantador = p.UserSession.Nome
		}
	}

	optionsJoined := strings.Join(cfg.Options, ",")
	modeStr := string(cfg.Mode)
	if modeStr == "" {
		modeStr = historydb.ModeCompleta
	}

	clientID, err := p.History.EnsureClient(cfg.ClientCode, "")
	if err != nil {
		p.OnLog(fmt.Sprintf("Erro ao registrar cliente no histórico: %v", err))
		return nil, err
	}
	runID, err := p.History.StartRun(clientID, cfg.System, optionsJoined, modeStr, implantador)
	if err != nil {
		p.OnLog(fmt.Sprintf("Erro ao iniciar run no histórico: %v", err))
		return nil, err
	}

	hasAnyError := false
	defer func() {
		if runID == "" || p.History == nil || !p.History.Connected() {
			return
		}
		st := historydb.RunStatusCompleto
		if p.Ctx.Err() != nil {
			st = historydb.RunStatusCancelado
		} else if hasAnyError {
			st = historydb.RunStatusParcial
		}
		_ = p.History.FinishRun(runID, st)
	}()

	runCtx := &services.MigrationRunContext{
		RunID:       runID,
		ClientID:    clientID,
		Implantador: implantador,
	}

	total := len(cfg.Options)
	for i, option := range cfg.Options {
		select {
		case <-p.Ctx.Done():
			p.OnLog("Migração cancelada pelo usuário")
			return res, p.Ctx.Err()
		default:
		}

		startTime := time.Now()
		status := models.MigrationStatus{
			Option:    option,
			Name:      p.Migration.GetOptionName(option, cfg.System),
			Status:    "running",
			StartTime: startTime.Format("15:04:05"),
		}
		if p.OnOptionStatus != nil {
			p.OnOptionStatus(i, status)
		}

		p.OnLog(fmt.Sprintf("Iniciando migração: %s", status.Name))
		logger.SetLogContext(cfg.ClientCode, cfg.System, option)

		stats, mErr := p.Migration.MigrateOption(p.Ctx, cfg, option, runCtx, func(msg string) {
			p.OnLog(msg)
		})

		endTime := time.Now()
		duration := endTime.Sub(startTime)
		status.EndTime = endTime.Format("15:04:05")
		status.Duration = duration.String()

		if stats != nil {
			status.TotalOrigem = stats.TotalOrigem
			status.Skipped = stats.Skipped
			status.Novos = stats.Novos
			status.ErrosReg = stats.Erros
		}

		if mErr != nil {
			hasAnyError = true
			res.HasPartialErrors = true
			status.Status = "error"
			status.Error = mErr.Error()
			p.OnLog(fmt.Sprintf("Erro ao migrar %s: %v", status.Name, mErr))
		} else {
			status.Status = "success"
			p.OnLog(fmt.Sprintf("%s concluído em %s", status.Name, duration.String()))
		}

		if p.OnOptionStatus != nil {
			p.OnOptionStatus(i, status)
		}

		if p.OnProgress != nil {
			p.OnProgress(i+1, total)
		}
	}

	p.OnLog("Migração concluída!")
	return res, nil
}
