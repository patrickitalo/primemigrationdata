package config

import (
	"github.com/joho/godotenv"
)

// LoadDotEnv carrega variáveis de um arquivo .env no diretório de trabalho atual.
// Se o arquivo não existir, não retorna erro (comportamento típico em dev/produção).
func LoadDotEnv() {
	_ = godotenv.Load()
}
