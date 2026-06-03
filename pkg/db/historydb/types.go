package historydb

// Status de execução em MIGRATION_RUNS.
const (
	RunStatusEmAndamento = "EM_ANDAMENTO"
	RunStatusCompleto    = "COMPLETO"
	RunStatusParcial     = "PARCIAL"
	RunStatusErro        = "ERRO"
	RunStatusCancelado   = "CANCELADO"
)

// Modo de migração.
const (
	ModeCompleta     = "COMPLETA"
	ModeIncremental  = "INCREMENTAL"
)
