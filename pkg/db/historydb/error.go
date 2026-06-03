package historydb

import (
	"database/sql"
	"fmt"
	"strings"
)

// InsertError grava erro granular por registro.
func InsertError(db *sql.DB, runID string, entityType int, sourceID, errorMsg string) error {
	if runID == "" {
		return fmt.Errorf("historydb: InsertError: runID vazio")
	}
	_, err := db.Exec(`
INSERT INTO MIGRATION_ERRORS (RUN_ID, ENTITY_TYPE, SOURCE_ID, ERROR_MSG)
VALUES (?, ?, ?, ?)`,
		runID, entityType, nullStr(strings.TrimSpace(sourceID)), errorMsg,
	)
	if err != nil {
		return fmt.Errorf("historydb: InsertError: %w", err)
	}
	return nil
}
