package logger

import (
	"log"
	"os"
	"strings"
	"sync"
)

var (
	fileLogMu   sync.Mutex
	fileLogFile *os.File
	fileLogPath string
)

// ConfigureFileFromEnv abre o arquivo de log em modo append.
// PRIME_LOG_FILE: caminho completo (ex.: D:/logs/prime-migration.log). Vazio = ./migration.log
// PRIME_DISABLE_FILE_LOG=1: não grava em arquivo.
func ConfigureFileFromEnv() {
	if strings.TrimSpace(os.Getenv("PRIME_DISABLE_FILE_LOG")) == "1" {
		return
	}
	path := strings.TrimSpace(os.Getenv("PRIME_LOG_FILE"))
	if path == "" {
		path = "migration.log"
	}
	if err := openSessionLogFile(path); err != nil {
		log.Printf("logger: arquivo de log não habilitado (%s): %v", path, err)
	}
}

// openSessionLogFile define o arquivo de sessão (reabre se o caminho mudar).
func openSessionLogFile(path string) error {
	fileLogMu.Lock()
	defer fileLogMu.Unlock()
	if fileLogFile != nil && fileLogPath == path {
		return nil
	}
	if fileLogFile != nil {
		_ = fileLogFile.Close()
		fileLogFile = nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fileLogPath = path
	fileLogFile = f
	return nil
}

func appendSessionFileLog(line string) {
	fileLogMu.Lock()
	defer fileLogMu.Unlock()
	if fileLogFile == nil {
		return
	}
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	_, _ = fileLogFile.WriteString(line)
	_ = fileLogFile.Sync()
}

// AppendSessionUILog grava uma linha já formatada (ex.: log da janela de migração).
func AppendSessionUILog(line string) {
	appendSessionFileLog(line)
}
