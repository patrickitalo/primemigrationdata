package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/nakagami/firebirdsql"
)

func Connect(cfg DbConfig) (*sql.DB, error) {
	// Validar configuração antes de tentar conectar
	if err := validarConfiguracao(cfg); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	connStr := fmt.Sprintf("%s:%s@%s:%s/%s?charset=WIN1252",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Path)

	db, err := sql.Open("firebirdsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com banco: %w", err)
	}

	// Configurar timeouts
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Testar conexão com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao testar conexão com banco: %w", err)
	}

	return db, nil
}

// validarConfiguracao valida se a configuração está correta
func validarConfiguracao(cfg DbConfig) error {
	if cfg.Path == "" {
		return fmt.Errorf("caminho do banco de dados não pode estar vazio")
	}
	if cfg.User == "" {
		return fmt.Errorf("usuário não pode estar vazio")
	}
	if cfg.Password == "" {
		return fmt.Errorf("senha não pode estar vazia")
	}
	if cfg.Host == "" {
		return fmt.Errorf("host não pode estar vazio")
	}
	if cfg.Port == "" {
		return fmt.Errorf("porta não pode estar vazia")
	}
	return nil
}
