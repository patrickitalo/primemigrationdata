package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const envHistoryDSN = "PRIME_HISTORY_DSN"

// centralFileName é o arquivo opcional no diretório de trabalho (não versionar credenciais).
const centralFileName = "central.json"

// CentralJSON formato mínimo do arquivo central.json.
type CentralJSON struct {
	HistoryDSN string `json:"history_dsn"`
}

// historyDSNFromEnvParts monta DSN a partir de PRIME_HISTORY_USER, PRIME_HISTORY_PASSWORD,
// PRIME_HISTORY_HOST, PRIME_HISTORY_PORT, PRIME_HISTORY_PATH, PRIME_HISTORY_CHARSET (opcional).
// Útil para manter senha só no .env sem colocá-la no central.json.
func historyDSNFromEnvParts() string {
	host := strings.TrimSpace(os.Getenv("PRIME_HISTORY_HOST"))
	path := strings.TrimSpace(os.Getenv("PRIME_HISTORY_PATH"))
	if host == "" || path == "" {
		return ""
	}
	user := strings.TrimSpace(os.Getenv("PRIME_HISTORY_USER"))
	if user == "" {
		user = "SYSDBA"
	}
	pass := os.Getenv("PRIME_HISTORY_PASSWORD")
	port := strings.TrimSpace(os.Getenv("PRIME_HISTORY_PORT"))
	if port == "" {
		port = "3050"
	}
	charset := strings.TrimSpace(os.Getenv("PRIME_HISTORY_CHARSET"))
	if charset == "" {
		charset = "WIN1252"
	}
	userInfo := url.UserPassword(user, pass)
	return fmt.Sprintf("%s@%s:%s/%s?charset=%s", userInfo.String(), host, port, path, charset)
}

// LoadHistoryDSN resolve DSN do Firebird central na ordem:
// PRIME_HISTORY_DSN → partes PRIME_HISTORY_* no .env → central.json → flag --history-dsn.
func LoadHistoryDSN(flagDSN string) string {
	if v := strings.TrimSpace(os.Getenv(envHistoryDSN)); v != "" {
		return v
	}
	if v := historyDSNFromEnvParts(); v != "" {
		return v
	}
	if v := loadCentralJSON(); v != "" {
		return v
	}
	return strings.TrimSpace(flagDSN)
}

func loadCentralJSON() string {
	data, err := os.ReadFile(centralFileName)
	if err != nil || len(data) == 0 {
		return ""
	}
	var c CentralJSON
	if err := json.Unmarshal(data, &c); err != nil {
		return ""
	}
	return strings.TrimSpace(c.HistoryDSN)
}
