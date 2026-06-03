package historydb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/nakagami/firebirdsql"
)

// Open abre conexão com o Firebird central usando DSN no formato nakagami/firebirdsql
// (ex.: user:pass@host:port/path/file.fdb?charset=WIN1252).
func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN do histórico vazio")
	}

	db, err := sql.Open("firebirdsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("historydb: abrir conexão: %w", err)
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("historydb: ping: %w", err)
	}

	return db, nil
}
