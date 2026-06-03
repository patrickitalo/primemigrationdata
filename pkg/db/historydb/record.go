package historydb

import (
	"database/sql"
	"fmt"
	"strings"
)

// RecordMeta destino e hash conhecidos no histórico para um SOURCE_ID.
type RecordMeta struct {
	DestinationID string
	SourceHash    string
}

// GetMigratedIDs retorna mapa SOURCE_ID -> DESTINATION_ID para o cliente e tipo de entidade.
func GetMigratedIDs(db *sql.DB, clientID, entityType int) (map[string]string, error) {
	meta, err := GetMigratedRecordMeta(db, clientID, entityType)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(meta))
	for k, v := range meta {
		out[k] = v.DestinationID
	}
	return out, nil
}

// GetMigratedRecordMeta retorna SOURCE_ID -> metadados para comparação incremental (hash).
func GetMigratedRecordMeta(db *sql.DB, clientID, entityType int) (map[string]RecordMeta, error) {
	if clientID <= 0 {
		return map[string]RecordMeta{}, nil
	}
	rows, err := db.Query(`
SELECT SOURCE_ID, DESTINATION_ID, SOURCE_HASH FROM MIGRATION_RECORDS
WHERE CLIENT_ID = ? AND ENTITY_TYPE = ?`,
		clientID, entityType,
	)
	if err != nil {
		return nil, fmt.Errorf("historydb: GetMigratedRecordMeta: %w", err)
	}
	defer rows.Close()

	out := make(map[string]RecordMeta)
	for rows.Next() {
		var src, dst, hsh sql.NullString
		if err := rows.Scan(&src, &dst, &hsh); err != nil {
			return nil, err
		}
		if !src.Valid {
			continue
		}
		key := strings.TrimSpace(src.String)
		m := RecordMeta{}
		if dst.Valid {
			m.DestinationID = strings.TrimSpace(dst.String)
		}
		if hsh.Valid {
			m.SourceHash = strings.TrimSpace(hsh.String)
		}
		out[key] = m
	}
	return out, rows.Err()
}

// UpsertRecord insere ou atualiza mapeamento (único por CLIENT_ID+ENTITY+SOURCE).
// Atualiza RUN_ID para o run corrente e hash quando já existia linha.
func UpsertRecord(db *sql.DB, runID string, clientID, entityType int, sourceID, destinationID, sourceHash string) error {
	if runID == "" || clientID <= 0 || strings.TrimSpace(sourceID) == "" {
		return fmt.Errorf("historydb: UpsertRecord: parâmetros inválidos")
	}
	src := strings.TrimSpace(sourceID)
	var n int
	err := db.QueryRow(`
SELECT COUNT(*) FROM MIGRATION_RECORDS WHERE CLIENT_ID = ? AND ENTITY_TYPE = ? AND SOURCE_ID = ?`,
		clientID, entityType, src,
	).Scan(&n)
	if err != nil {
		return fmt.Errorf("historydb: UpsertRecord select: %w", err)
	}
	if n == 0 {
		_, err = db.Exec(`
INSERT INTO MIGRATION_RECORDS (RUN_ID, CLIENT_ID, ENTITY_TYPE, SOURCE_ID, DESTINATION_ID, SOURCE_HASH)
VALUES (?, ?, ?, ?, ?, ?)`,
			runID, clientID, entityType, src, nullStr(destinationID), nullStr(sourceHash),
		)
		if err != nil {
			return fmt.Errorf("historydb: UpsertRecord insert: %w", err)
		}
		return nil
	}
	_, err = db.Exec(`
UPDATE MIGRATION_RECORDS SET RUN_ID = ?, DESTINATION_ID = ?, SOURCE_HASH = ?, MIGRATED_AT = CURRENT_TIMESTAMP
WHERE CLIENT_ID = ? AND ENTITY_TYPE = ? AND SOURCE_ID = ?`,
		runID, nullStr(destinationID), nullStr(sourceHash), clientID, entityType, src,
	)
	if err != nil {
		return fmt.Errorf("historydb: UpsertRecord update: %w", err)
	}
	return nil
}
