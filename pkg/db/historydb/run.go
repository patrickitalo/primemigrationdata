package historydb

import (
	"database/sql"
	"fmt"
	"time"
)

// MigrationRunRow espelha uma linha útil para consultas.
type MigrationRunRow struct {
	ID          string
	ClientID    int
	System      string
	Options     string
	Mode        string
	Status      string
	StartedAt   time.Time
	FinishedAt  sql.NullTime
	Implantador string
}

// InsertRun insere um novo run em EM_ANDAMENTO.
func InsertRun(db *sql.DB, runID string, clientID int, system, options, mode, implantador string) error {
	if runID == "" || clientID <= 0 {
		return fmt.Errorf("historydb: InsertRun: parâmetros inválidos")
	}
	if mode == "" {
		mode = ModeCompleta
	}
	_, err := db.Exec(`
INSERT INTO MIGRATION_RUNS (ID, CLIENT_ID, SYSTEM, OPTIONS, MODE, STATUS, IMPLANTADOR)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID, clientID, system, nullStr(options), mode, RunStatusEmAndamento, nullStr(implantador),
	)
	if err != nil {
		return fmt.Errorf("historydb: InsertRun: %w", err)
	}
	return nil
}

// FinishRun atualiza status e FINISHED_AT.
func FinishRun(db *sql.DB, runID, status string) error {
	if runID == "" || status == "" {
		return fmt.Errorf("historydb: FinishRun: parâmetros inválidos")
	}
	_, err := db.Exec(
		`UPDATE MIGRATION_RUNS SET STATUS = ?, FINISHED_AT = CURRENT_TIMESTAMP WHERE ID = ?`,
		status, runID,
	)
	if err != nil {
		return fmt.Errorf("historydb: FinishRun: %w", err)
	}
	return nil
}

// GetLastSuccessfulRun retorna o último run COMPLETO para cliente e sistema.
func GetLastSuccessfulRun(db *sql.DB, clientID int, system string) (*MigrationRunRow, error) {
	if clientID <= 0 {
		return nil, nil
	}
	var r MigrationRunRow
	var finished sql.NullTime
	err := db.QueryRow(`
SELECT FIRST 1 ID, CLIENT_ID, SYSTEM, OPTIONS, MODE, STATUS, STARTED_AT, FINISHED_AT, IMPLANTADOR
FROM MIGRATION_RUNS
WHERE CLIENT_ID = ? AND SYSTEM = ? AND STATUS = ?
ORDER BY STARTED_AT DESC`,
		clientID, system, RunStatusCompleto,
	).Scan(&r.ID, &r.ClientID, &r.System, &r.Options, &r.Mode, &r.Status, &r.StartedAt, &finished, &r.Implantador)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("historydb: GetLastSuccessfulRun: %w", err)
	}
	r.FinishedAt = finished
	return &r, nil
}
