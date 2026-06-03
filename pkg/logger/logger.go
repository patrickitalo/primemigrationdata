package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Variáveis para armazenar contexto de logging (cliente, sistema, tipo de migração)
var (
	logContextMu       sync.RWMutex
	currentCodigoCliente string
	currentSistemaOrigem  string
	currentTipoMigracao   string
	
	// Hook de callback para salvar logs no SQLite (configurado externamente)
	logSQLiteHook func(level, message, codigoCliente, sistemaOrigem, tipoMigracao string)
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	level LogLevel
	*log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger(INFO)
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	levelStr := ""
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	case FATAL:
		levelStr = "FATAL"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog(DEBUG) {
		message := fmt.Sprintf(format, args...)
		line := l.formatMessage(DEBUG, format, args...)
		l.Printf("%s", line)
		appendSessionFileLog(line + "\n")
		saveToSQLite(DEBUG, message)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog(INFO) {
		message := fmt.Sprintf(format, args...)
		line := l.formatMessage(INFO, format, args...)
		l.Printf("%s", line)
		appendSessionFileLog(line + "\n")
		saveToSQLite(INFO, message)
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog(WARN) {
		message := fmt.Sprintf(format, args...)
		line := l.formatMessage(WARN, format, args...)
		l.Printf("%s", line)
		appendSessionFileLog(line + "\n")
		saveToSQLite(WARN, message)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog(ERROR) {
		message := fmt.Sprintf(format, args...)
		line := l.formatMessage(ERROR, format, args...)
		l.Printf("%s", line)
		appendSessionFileLog(line + "\n")
		saveToSQLite(ERROR, message)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.shouldLog(FATAL) {
		message := fmt.Sprintf(format, args...)
		line := l.formatMessage(FATAL, format, args...)
		l.Printf("%s", line)
		appendSessionFileLog(line + "\n")
		saveToSQLite(FATAL, message)
		os.Exit(1)
	}
}

// Funções globais para facilitar o uso
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetLogContext define o contexto atual para logs (cliente, sistema, tipo de migração)
func SetLogContext(codigoCliente, sistemaOrigem, tipoMigracao string) {
	logContextMu.Lock()
	defer logContextMu.Unlock()
	currentCodigoCliente = codigoCliente
	currentSistemaOrigem = sistemaOrigem
	currentTipoMigracao = tipoMigracao
}

// GetLogContext retorna o contexto atual de forma thread-safe (exportado para uso externo)
func GetLogContext() (codigoCliente, sistemaOrigem, tipoMigracao string) {
	logContextMu.RLock()
	defer logContextMu.RUnlock()
	return currentCodigoCliente, currentSistemaOrigem, currentTipoMigracao
}

// getLogContext retorna o contexto atual de forma thread-safe (interno)
func getLogContext() (codigoCliente, sistemaOrigem, tipoMigracao string) {
	return GetLogContext()
}

// SetLogSQLiteHook configura uma função callback para salvar logs no SQLite
func SetLogSQLiteHook(hook func(level, message, codigoCliente, sistemaOrigem, tipoMigracao string)) {
	logContextMu.Lock()
	defer logContextMu.Unlock()
	logSQLiteHook = hook
}

// saveToSQLite salva o log no SQLite de forma assíncrona (não bloqueia)
func saveToSQLite(level LogLevel, message string) {
	// Usar goroutine para não bloquear
	go func() {
		defer func() {
			// Recuperar qualquer panic para não quebrar o fluxo
			if r := recover(); r != nil {
				// Silenciosamente ignorar erros de salvamento
			}
		}()

		logContextMu.RLock()
		hook := logSQLiteHook
		codigoCliente := currentCodigoCliente
		sistemaOrigem := currentSistemaOrigem
		tipoMigracao := currentTipoMigracao
		logContextMu.RUnlock()

		// Se não há hook configurado, não fazer nada
		if hook == nil {
			return
		}

		levelStr := ""
		switch level {
		case DEBUG:
			levelStr = "DEBUG"
		case INFO:
			levelStr = "INFO"
		case WARN:
			levelStr = "WARN"
		case ERROR:
			levelStr = "ERROR"
		case FATAL:
			levelStr = "FATAL"
		}

		// Chamar o hook para salvar no SQLite
		hook(levelStr, message, codigoCliente, sistemaOrigem, tipoMigracao)
	}()
}
