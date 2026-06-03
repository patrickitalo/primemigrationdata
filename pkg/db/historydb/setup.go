package historydb

import (
	"database/sql"
	"fmt"
	"strings"
)

// isSafeHistoryMetaName restringe nomes embutidos em SQL de catálogo (evita injeção).
func isSafeHistoryMetaName(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func sqlStringLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// Setup cria generators, tabelas, índices e triggers se ainda não existirem (Firebird 2.5).
func Setup(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("historydb: db nulo")
	}

	steps := []func(*sql.DB) error{
		ensureGenerator("GEN_CLIENTES_MIGRACAO_ID"),
		ensureGenerator("GEN_MIGRATION_RECORDS_ID"),
		ensureGenerator("GEN_MIGRATION_ERRORS_ID"),
		ensureTableClientesMigracao,
		ensureTableMigrationRuns,
		ensureTableMigrationRecords,
		ensureTableMigrationErrors,
		ensureIndexRecordsUnique,
	}

	for _, step := range steps {
		if err := step(db); err != nil {
			return err
		}
	}
	return nil
}

func generatorExists(db *sql.DB, name string) (bool, error) {
	if !isSafeHistoryMetaName(name) {
		return false, fmt.Errorf("historydb: nome de generator inválido")
	}
	var n int
	q := `SELECT COUNT(*) FROM RDB$GENERATORS WHERE TRIM(RDB$GENERATOR_NAME) = ` + sqlStringLiteral(name)
	err := db.QueryRow(q).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func ensureGenerator(name string) func(*sql.DB) error {
	return func(db *sql.DB) error {
		ok, err := generatorExists(db, name)
		if err != nil {
			return fmt.Errorf("historydb: verificar generator %s: %w", name, err)
		}
		if ok {
			return nil
		}
		_, err = db.Exec(fmt.Sprintf("CREATE GENERATOR %s", name))
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "exists") {
			return fmt.Errorf("historydb: criar generator %s: %w", name, err)
		}
		_, _ = db.Exec(fmt.Sprintf("SET GENERATOR %s TO 0", name))
		return nil
	}
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	if !isSafeHistoryMetaName(tableName) {
		return false, fmt.Errorf("historydb: nome de tabela inválido")
	}
	var n int
	q := `SELECT COUNT(*) FROM RDB$RELATIONS WHERE RDB$SYSTEM_FLAG = 0 AND TRIM(RDB$RELATION_NAME) = ` + sqlStringLiteral(tableName)
	err := db.QueryRow(q).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func triggerExists(db *sql.DB, triggerName string) (bool, error) {
	if !isSafeHistoryMetaName(triggerName) {
		return false, fmt.Errorf("historydb: nome de trigger inválido")
	}
	var n int
	q := `SELECT COUNT(*) FROM RDB$TRIGGERS WHERE TRIM(RDB$TRIGGER_NAME) = ` + sqlStringLiteral(triggerName)
	err := db.QueryRow(q).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func ensureTableClientesMigracao(db *sql.DB) error {
	ok, err := tableExists(db, "CLIENTES_MIGRACAO")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = db.Exec(`
CREATE TABLE CLIENTES_MIGRACAO (
    ID INTEGER NOT NULL PRIMARY KEY,
    CLIENT_CODE VARCHAR(50) NOT NULL UNIQUE,
    NOME VARCHAR(200),
    CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		return fmt.Errorf("historydb: CLIENTES_MIGRACAO: %w", err)
	}
	tok, err := triggerExists(db, "BI_CLIENTES_MIGRACAO")
	if err != nil {
		return err
	}
	if !tok {
		_, err = db.Exec(`
CREATE TRIGGER BI_CLIENTES_MIGRACAO FOR CLIENTES_MIGRACAO
ACTIVE BEFORE INSERT POSITION 0
AS
BEGIN
  IF (NEW.ID IS NULL OR NEW.ID = 0) THEN
    NEW.ID = GEN_ID(GEN_CLIENTES_MIGRACAO_ID, 1);
END`)
		if err != nil {
			return fmt.Errorf("historydb: trigger CLIENTES_MIGRACAO: %w", err)
		}
	}
	return nil
}

func ensureTableMigrationRuns(db *sql.DB) error {
	ok, err := tableExists(db, "MIGRATION_RUNS")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = db.Exec(`
CREATE TABLE MIGRATION_RUNS (
    ID CHAR(36) NOT NULL PRIMARY KEY,
    CLIENT_ID INTEGER NOT NULL REFERENCES CLIENTES_MIGRACAO(ID),
    SYSTEM VARCHAR(20) NOT NULL,
    OPTIONS VARCHAR(30),
    MODE VARCHAR(20) DEFAULT 'COMPLETA',
    STATUS VARCHAR(20) DEFAULT 'EM_ANDAMENTO',
    STARTED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FINISHED_AT TIMESTAMP,
    IMPLANTADOR VARCHAR(100),
    OBS BLOB SUB_TYPE TEXT
)`)
	if err != nil {
		return fmt.Errorf("historydb: MIGRATION_RUNS: %w", err)
	}
	return nil
}

func ensureTableMigrationRecords(db *sql.DB) error {
	ok, err := tableExists(db, "MIGRATION_RECORDS")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = db.Exec(`
CREATE TABLE MIGRATION_RECORDS (
    ID INTEGER NOT NULL PRIMARY KEY,
    RUN_ID CHAR(36) NOT NULL REFERENCES MIGRATION_RUNS(ID),
    CLIENT_ID INTEGER NOT NULL REFERENCES CLIENTES_MIGRACAO(ID),
    ENTITY_TYPE SMALLINT NOT NULL,
    SOURCE_ID VARCHAR(50) NOT NULL,
    DESTINATION_ID VARCHAR(50),
    SOURCE_HASH VARCHAR(64),
    MIGRATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		return fmt.Errorf("historydb: MIGRATION_RECORDS: %w", err)
	}
	tok, err := triggerExists(db, "BI_MIGRATION_RECORDS")
	if err != nil {
		return err
	}
	if !tok {
		_, err = db.Exec(`
CREATE TRIGGER BI_MIGRATION_RECORDS FOR MIGRATION_RECORDS
ACTIVE BEFORE INSERT POSITION 0
AS
BEGIN
  IF (NEW.ID IS NULL OR NEW.ID = 0) THEN
    NEW.ID = GEN_ID(GEN_MIGRATION_RECORDS_ID, 1);
END`)
		if err != nil {
			return fmt.Errorf("historydb: trigger MIGRATION_RECORDS: %w", err)
		}
	}
	return nil
}

func ensureTableMigrationErrors(db *sql.DB) error {
	ok, err := tableExists(db, "MIGRATION_ERRORS")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = db.Exec(`
CREATE TABLE MIGRATION_ERRORS (
    ID INTEGER NOT NULL PRIMARY KEY,
    RUN_ID CHAR(36) NOT NULL REFERENCES MIGRATION_RUNS(ID),
    ENTITY_TYPE SMALLINT,
    SOURCE_ID VARCHAR(50),
    ERROR_MSG BLOB SUB_TYPE TEXT,
    OCCURRED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		return fmt.Errorf("historydb: MIGRATION_ERRORS: %w", err)
	}
	tok, err := triggerExists(db, "BI_MIGRATION_ERRORS")
	if err != nil {
		return err
	}
	if !tok {
		_, err = db.Exec(`
CREATE TRIGGER BI_MIGRATION_ERRORS FOR MIGRATION_ERRORS
ACTIVE BEFORE INSERT POSITION 0
AS
BEGIN
  IF (NEW.ID IS NULL OR NEW.ID = 0) THEN
    NEW.ID = GEN_ID(GEN_MIGRATION_ERRORS_ID, 1);
END`)
		if err != nil {
			return fmt.Errorf("historydb: trigger MIGRATION_ERRORS: %w", err)
		}
	}
	return nil
}

func indexExists(db *sql.DB, indexName string) (bool, error) {
	if !isSafeHistoryMetaName(indexName) {
		return false, fmt.Errorf("historydb: nome de índice inválido")
	}
	var n int
	q := `SELECT COUNT(*) FROM RDB$INDICES WHERE TRIM(RDB$INDEX_NAME) = ` + sqlStringLiteral(indexName) + ` AND RDB$SYSTEM_FLAG = 0`
	err := db.QueryRow(q).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func ensureIndexRecordsUnique(db *sql.DB) error {
	// Nome ≤31 caracteres (limite Firebird 2.5).
	ok, err := indexExists(db, "UQ_MIGREC_CLI_ENT_SRC")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = db.Exec(`
CREATE UNIQUE INDEX UQ_MIGREC_CLI_ENT_SRC
ON MIGRATION_RECORDS (CLIENT_ID, ENTITY_TYPE, SOURCE_ID)`)
	if err != nil {
		return fmt.Errorf("historydb: índice records: %w", err)
	}
	return nil
}
