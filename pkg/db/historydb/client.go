package historydb

import (
	"database/sql"
	"fmt"
	"strings"
)

// GetClientID retorna o ID interno do cliente ou 0 se não existir.
func GetClientID(db *sql.DB, clientCode string) (int, error) {
	var id int
	err := db.QueryRow(
		`SELECT ID FROM CLIENTES_MIGRACAO WHERE CLIENT_CODE = ?`,
		strings.TrimSpace(clientCode),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("historydb: GetClientID: %w", err)
	}
	return id, nil
}

// UpsertClient insere ou atualiza nome do cliente e retorna o ID.
func UpsertClient(db *sql.DB, clientCode, nome string) (int, error) {
	code := strings.TrimSpace(clientCode)
	if code == "" {
		return 0, fmt.Errorf("historydb: client_code vazio")
	}

	id, err := GetClientID(db, code)
	if err != nil {
		return 0, err
	}
	if id > 0 {
		if strings.TrimSpace(nome) != "" {
			_, err = db.Exec(
				`UPDATE CLIENTES_MIGRACAO SET NOME = ?, UPDATED_AT = CURRENT_TIMESTAMP WHERE ID = ?`,
				nome, id,
			)
			if err != nil {
				return 0, fmt.Errorf("historydb: atualizar cliente: %w", err)
			}
		}
		return id, nil
	}

	_, err = db.Exec(
		`INSERT INTO CLIENTES_MIGRACAO (CLIENT_CODE, NOME) VALUES (?, ?)`,
		code, nullStr(nome),
	)
	if err != nil {
		return 0, fmt.Errorf("historydb: inserir cliente: %w", err)
	}

	return GetClientID(db, code)
}

func nullStr(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}
