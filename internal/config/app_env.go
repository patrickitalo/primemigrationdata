package config

import (
	"os"
	"strings"
)

// FirebirdFormEnv valores opcionais para preencher a UI (origem Firebird).
type FirebirdFormEnv struct {
	Host, Port, Path, User, Password, Conversao string
}

// FirebirdFormEnvFromOS lê PRIME_FIREBIRD_* do ambiente (incluindo após LoadDotEnv).
func FirebirdFormEnvFromOS() FirebirdFormEnv {
	return FirebirdFormEnv{
		Host:      strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_HOST")),
		Port:      strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_PORT")),
		Path:      strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_PATH")),
		User:      strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_USER")),
		Password:  os.Getenv("PRIME_FIREBIRD_PASSWORD"),
		Conversao: strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_CONVERSAO")),
	}
}

// PharmacieFormEnv valores opcionais para conexão Pharmacie (destino).
type PharmacieFormEnv struct {
	Alias, IPServer, Porta string
}

// PharmacieFormEnvFromOS lê PRIME_PHARMACIE_*.
func PharmacieFormEnvFromOS() PharmacieFormEnv {
	return PharmacieFormEnv{
		Alias:    strings.TrimSpace(os.Getenv("PRIME_PHARMACIE_ALIAS")),
		IPServer: strings.TrimSpace(os.Getenv("PRIME_PHARMACIE_IPSERVER")),
		Porta:    strings.TrimSpace(os.Getenv("PRIME_PHARMACIE_PORT")),
	}
}

// EffectiveFirebirdPassword usa PRIME_FIREBIRD_PASSWORD do ambiente quando definida;
// caso contrário usa o valor persistido no SQLite.
func EffectiveFirebirdPassword(savedInSQLite string) string {
	if strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_PASSWORD")) != "" {
		return os.Getenv("PRIME_FIREBIRD_PASSWORD")
	}
	return savedInSQLite
}

// ShouldOmitFirebirdPasswordOnSave evita gravar senha no config.db quando ela vem só do .env.
func ShouldOmitFirebirdPasswordOnSave() bool {
	return strings.TrimSpace(os.Getenv("PRIME_FIREBIRD_PASSWORD")) != ""
}
