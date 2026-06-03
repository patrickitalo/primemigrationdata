package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/primesoftwaresi/prime-migration/pkg/db/historydb"
)

// HistoryService fachada para o Firebird central de histórico de migrações.
type HistoryService struct {
	mu        sync.RWMutex
	db        *sql.DB
	connected bool
	lastErr   error
}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

// Connected indica se há conexão ativa com o banco central.
func (h *HistoryService) Connected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.connected && h.db != nil
}

// LastError retorna o último erro de conexão/setup (se houver).
func (h *HistoryService) LastError() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastErr
}

// Connect abre o DSN e aplica Setup (DDL idempotente).
func (h *HistoryService) Connect(dsn string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastErr = nil
	if strings.TrimSpace(dsn) == "" {
		h.connected = false
		err := fmt.Errorf("DSN do histórico não configurado (.env, PRIME_HISTORY_DSN, partes PRIME_HISTORY_*, central.json ou --history-dsn)")
		h.lastErr = err
		return err
	}

	if h.db != nil {
		_ = h.db.Close()
		h.db = nil
	}

	db, err := historydb.Open(dsn)
	if err != nil {
		h.connected = false
		h.lastErr = err
		return err
	}
	if err := historydb.Setup(db); err != nil {
		_ = db.Close()
		h.connected = false
		h.lastErr = err
		return err
	}

	h.db = db
	h.connected = true
	return nil
}

// Close encerra a conexão com o banco central.
func (h *HistoryService) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.db == nil {
		return nil
	}
	err := h.db.Close()
	h.db = nil
	h.connected = false
	return err
}

func (h *HistoryService) dbOrErr() (*sql.DB, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.db == nil || !h.connected {
		return nil, fmt.Errorf("histórico central não conectado")
	}
	return h.db, nil
}

// EnsureClient registra o cliente e retorna o ID interno.
func (h *HistoryService) EnsureClient(clientCode, nome string) (int, error) {
	db, err := h.dbOrErr()
	if err != nil {
		return 0, err
	}
	return historydb.UpsertClient(db, clientCode, nome)
}

// StartRun cria um run e retorna o UUID.
func (h *HistoryService) StartRun(clientID int, system, options, mode, implantador string) (string, error) {
	db, err := h.dbOrErr()
	if err != nil {
		return "", err
	}
	if mode == "" {
		mode = historydb.ModeCompleta
	}
	runID := newRunIDv4()
	if err := historydb.InsertRun(db, runID, clientID, system, options, mode, implantador); err != nil {
		return "", err
	}
	return runID, nil
}

// FinishRun finaliza o run com o status informado.
func (h *HistoryService) FinishRun(runID, status string) error {
	db, err := h.dbOrErr()
	if err != nil {
		return err
	}
	return historydb.FinishRun(db, runID, status)
}

// SaveRecord grava ou atualiza mapeamento source→destino.
func (h *HistoryService) SaveRecord(runID string, clientID, entityType int, sourceID, destinationID, hash string) error {
	db, err := h.dbOrErr()
	if err != nil {
		return err
	}
	return historydb.UpsertRecord(db, runID, clientID, entityType, sourceID, destinationID, hash)
}

// SaveError grava erro por registro.
func (h *HistoryService) SaveError(runID string, entityType int, sourceID, errorMsg string) error {
	db, err := h.dbOrErr()
	if err != nil {
		return err
	}
	return historydb.InsertError(db, runID, entityType, sourceID, errorMsg)
}

// GetMigratedIDs retorna source_id → destination_id já migrados.
func (h *HistoryService) GetMigratedIDs(clientID, entityType int) (map[string]string, error) {
	db, err := h.dbOrErr()
	if err != nil {
		return nil, err
	}
	return historydb.GetMigratedIDs(db, clientID, entityType)
}

// GetMigratedRecordMeta retorna source_id → metadados (destino e hash) para modo incremental.
func (h *HistoryService) GetMigratedRecordMeta(clientID, entityType int) (map[string]historydb.RecordMeta, error) {
	db, err := h.dbOrErr()
	if err != nil {
		return nil, err
	}
	return historydb.GetMigratedRecordMeta(db, clientID, entityType)
}

// MigrationRun informações do último run bem-sucedido (uso na UI).
type MigrationRun struct {
	RunID       string
	System      string
	Options     string
	Mode        string
	Status      string
	StartedAt   time.Time
	FinishedAt  *time.Time
	Implantador string
}

// GetLastSuccessfulRun consulta o último run COMPLETO para cliente e sistema.
func (h *HistoryService) GetLastSuccessfulRun(clientID int, system string) (*MigrationRun, error) {
	db, err := h.dbOrErr()
	if err != nil {
		return nil, err
	}
	row, err := historydb.GetLastSuccessfulRun(db, clientID, system)
	if err != nil || row == nil {
		return nil, err
	}
	m := &MigrationRun{
		RunID:       row.ID,
		System:      row.System,
		Options:     row.Options,
		Mode:        row.Mode,
		Status:      row.Status,
		StartedAt:   row.StartedAt,
		Implantador: row.Implantador,
	}
	if row.FinishedAt.Valid {
		t := row.FinishedAt.Time
		m.FinishedAt = &t
	}
	return m, nil
}

func newRunIDv4() string {
	var u [16]byte
	_, _ = rand.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80
	h := hex.EncodeToString(u[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}
