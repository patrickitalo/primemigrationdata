package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/primesoftwaresi/prime-migration-fyne/internal/config"
	"github.com/primesoftwaresi/prime-migration-fyne/internal/models"
	"github.com/primesoftwaresi/prime-migration-fyne/internal/services"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const AppVersion = "2.0.4"

// App is the Wails application struct. All exported methods become TypeScript bindings.
type App struct {
	ctx              context.Context
	authService      *services.AuthService
	configService    *services.ConfigService
	historyService   *services.HistoryService
	migrationService *services.MigrationService
	historyDSN       string
	session          *models.UserSession

	migMu     sync.Mutex
	migCancel context.CancelFunc
}

func NewApp(
	authService *services.AuthService,
	configService *services.ConfigService,
	historyService *services.HistoryService,
	migrationService *services.MigrationService,
	historyDSN string,
) *App {
	return &App{
		authService:      authService,
		configService:    configService,
		historyService:   historyService,
		migrationService: migrationService,
		historyDSN:       historyDSN,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(_ context.Context) {
	a.historyService.Close()
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

type LoginResult struct {
	Session *models.UserSession `json:"session,omitempty"`
	Error   string              `json:"error,omitempty"`
}

func (a *App) Login(username, password string) LoginResult {
	session, err := a.authService.Authenticate(username, password)
	if err != nil {
		return LoginResult{Error: err.Error()}
	}
	a.session = session
	if a.historyDSN != "" && !a.historyService.Connected() {
		if err := a.historyService.Connect(a.historyDSN); err == nil {
			runtime.EventsEmit(a.ctx, "history:status-changed", true)
		}
	}
	return LoginResult{Session: session}
}

func (a *App) Logout() {
	a.session = nil
}

// ─── Environment defaults ──────────────────────────────────────────────────────

type EnvDefaults struct {
	Firebird  config.FirebirdFormEnv  `json:"firebird"`
	Pharmacie config.PharmacieFormEnv `json:"pharmacie"`
}

func (a *App) GetEnvDefaults() EnvDefaults {
	return EnvDefaults{
		Firebird:  config.FirebirdFormEnvFromOS(),
		Pharmacie: config.PharmacieFormEnvFromOS(),
	}
}

// ─── Client config ─────────────────────────────────────────────────────────────

func (a *App) LoadClientConfig(clientCode, system string) (*models.ClientConfig, error) {
	return a.configService.LoadClientConfig(clientCode, system)
}

func (a *App) SaveClientConfig(cfg models.ClientConfig) error {
	return a.configService.SaveClientConfig(&cfg)
}

// ─── Migration options ──────────────────────────────────────────────────────────

type OptionDef struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (a *App) GetOptionNames(system string) []OptionDef {
	codes := []string{"1", "2", "3", "4", "5", "6", "7"}
	if system == "PRISMA5" {
		codes = append(codes, "8")
	}
	defs := make([]OptionDef, 0, len(codes))
	for _, c := range codes {
		defs = append(defs, OptionDef{Code: c, Name: a.migrationService.GetOptionName(c, system)})
	}
	return defs
}

func (a *App) GetAppVersion() string {
	return AppVersion
}

// ─── History DB ────────────────────────────────────────────────────────────────

type HistoryStatus struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

func (a *App) CheckHistoryStatus() HistoryStatus {
	if a.historyService.Connected() {
		return HistoryStatus{Connected: true}
	}
	errMsg := ""
	if err := a.historyService.LastError(); err != nil {
		errMsg = err.Error()
	}
	return HistoryStatus{Connected: false, Error: errMsg}
}

func (a *App) ReconnectHistory() error {
	if a.historyDSN == "" {
		return fmt.Errorf("DSN do histórico não configurado")
	}
	err := a.historyService.Connect(a.historyDSN)
	runtime.EventsEmit(a.ctx, "history:status-changed", err == nil)
	return err
}

// ─── Last run info ─────────────────────────────────────────────────────────────

type LastRunInfo struct {
	RunID       string `json:"runId"`
	Options     string `json:"options"`
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	StartedAt   string `json:"startedAt"`
	FinishedAt  string `json:"finishedAt,omitempty"`
	Implantador string `json:"implantador"`
}

func (a *App) GetLastRunInfo(clientCode, system string) *LastRunInfo {
	if !a.historyService.Connected() {
		return nil
	}
	clientID := 0
	if id, err := a.historyService.EnsureClient(clientCode, ""); err == nil {
		clientID = id
	}
	if clientID == 0 {
		return nil
	}
	run, err := a.historyService.GetLastSuccessfulRun(clientID, system)
	if err != nil || run == nil {
		return nil
	}
	info := &LastRunInfo{
		RunID:       run.RunID,
		Options:     run.Options,
		Mode:        run.Mode,
		Status:      run.Status,
		StartedAt:   run.StartedAt.Format("02/01/2006 15:04"),
		Implantador: run.Implantador,
	}
	if run.FinishedAt != nil {
		info.FinishedAt = run.FinishedAt.Format("02/01/2006 15:04")
	}
	return info
}

// ─── File dialog ───────────────────────────────────────────────────────────────

func (a *App) SelectExcelFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar Planilha Excel",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) SelectDBFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar Banco de Dados Firebird",
		Filters: []runtime.FileFilter{
			{DisplayName: "Firebird Database (*.fdb;*.gdb)", Pattern: "*.fdb;*.gdb"},
			{DisplayName: "Todos os arquivos (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// ─── Migration ──────────────────────────────────────────────────────────────────

type MigrationProgressEvent struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

type MigrationCompleteEvent struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (a *App) StartMigration(cfg models.MigrationConfig) error {
	a.migMu.Lock()
	defer a.migMu.Unlock()

	if a.migCancel != nil {
		return fmt.Errorf("migração já em andamento")
	}

	if err := a.configService.SaveClientConfig(&models.ClientConfig{
		CodigoCliente:     cfg.ClientCode,
		SistemaOrigem:     cfg.System,
		DbHost:            cfg.Database.Host,
		DbPort:            cfg.Database.Port,
		DbPath:            cfg.Database.Path,
		DbUser:            cfg.Database.User,
		DbPassword:        cfg.Database.Password,
		AliasPharmacie:    cfg.AliasPharmacie,
		IpServerPharmacie: cfg.IpServerPharmacie,
		PortaPharmacie:    cfg.PortaPharmacie,
		Conversao:         cfg.Database.Conversao,
	}); err != nil {
		a.emitLog(fmt.Sprintf("[AVISO] Falha ao salvar config: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.migCancel = cancel

	go a.runMigration(ctx, cfg)
	return nil
}

func (a *App) StopMigration() {
	a.migMu.Lock()
	defer a.migMu.Unlock()
	if a.migCancel != nil {
		a.migCancel()
	}
}

func (a *App) runMigration(ctx context.Context, cfg models.MigrationConfig) {
	defer func() {
		a.migMu.Lock()
		a.migCancel = nil
		a.migMu.Unlock()
	}()

	total := len(cfg.Options)
	completed := 0

	a.emitLog("Conectando ao banco de dados de origem...")

	database, err := a.migrationService.ConnectDatabase(cfg.Database)
	if err != nil {
		a.emitLog(fmt.Sprintf("[ERRO] Falha na conexão: %v", err))
		runtime.EventsEmit(a.ctx, "migration:complete", MigrationCompleteEvent{Success: false, Error: err.Error()})
		return
	}
	defer database.Close()

	a.emitLog("Configurando procedures...")
	if err := a.migrationService.SetupProcedures(database, cfg.System, cfg.ClientCode); err != nil {
		a.emitLog(fmt.Sprintf("[ERRO] Setup falhou: %v", err))
		runtime.EventsEmit(a.ctx, "migration:complete", MigrationCompleteEvent{Success: false, Error: err.Error()})
		return
	}

	if cfg.System == "PRISMA5" && cfg.ExcelPath != "" {
		a.emitLog("Importando planilha Excel...")
		if err := a.migrationService.ImportExcel(database, cfg.ExcelPath); err != nil {
			a.emitLog(fmt.Sprintf("[AVISO] Excel: %v", err))
		}
	}

	var runCtx *services.MigrationRunContext
	if a.historyService.Connected() && a.session != nil {
		clientID, _ := a.historyService.EnsureClient(cfg.ClientCode, cfg.ClientCode)
		optionsStr := ""
		for i, o := range cfg.Options {
			if i > 0 {
				optionsStr += ","
			}
			optionsStr += o
		}
		runID, err := a.historyService.StartRun(clientID, cfg.System, optionsStr, string(cfg.Mode), a.session.Nick)
		if err == nil {
			runCtx = &services.MigrationRunContext{RunID: runID, ClientID: clientID, Implantador: a.session.Nick}
		}
	}

	for _, option := range cfg.Options {
		if ctx.Err() != nil {
			a.emitLog("Migração cancelada pelo usuário.")
			break
		}

		optName := a.migrationService.GetOptionName(option, cfg.System)
		a.emitStatusUpdate(models.MigrationStatus{
			Option: option, Name: optName, Status: "running",
			StartTime: time.Now().Format("15:04:05"),
		})

		stats, err := a.migrationService.MigrateOption(ctx, &cfg, option, runCtx, func(msg string) {
			a.emitLog(msg)
		})

		status := models.MigrationStatus{
			Option: option, Name: optName,
			EndTime: time.Now().Format("15:04:05"),
		}
		if stats != nil {
			status.TotalOrigem = stats.TotalOrigem
			status.Skipped = stats.Skipped
			status.Novos = stats.Novos
			status.ErrosReg = stats.Erros
		}
		if err != nil {
			status.Status = "error"
			status.Error = err.Error()
			a.emitLog(fmt.Sprintf("[ERRO] %s: %v", optName, err))
		} else {
			status.Status = "success"
		}
		a.emitStatusUpdate(status)

		completed++
		runtime.EventsEmit(a.ctx, "migration:progress", MigrationProgressEvent{Completed: completed, Total: total})
	}

	if runCtx != nil {
		finalStatus := "success"
		if ctx.Err() != nil {
			finalStatus = "cancelled"
		}
		_ = a.historyService.FinishRun(runCtx.RunID, finalStatus)
	}

	if ctx.Err() != nil {
		runtime.EventsEmit(a.ctx, "migration:complete", MigrationCompleteEvent{Success: false, Error: "cancelado"})
		return
	}

	a.emitLog("Migração concluída com sucesso!")
	runtime.EventsEmit(a.ctx, "migration:complete", MigrationCompleteEvent{Success: true})
}

func (a *App) emitLog(msg string) {
	ts := time.Now().Format("15:04:05")
	runtime.EventsEmit(a.ctx, "migration:log", fmt.Sprintf("[%s] %s", ts, msg))
}

func (a *App) emitStatusUpdate(s models.MigrationStatus) {
	runtime.EventsEmit(a.ctx, "migration:status-update", s)
}
