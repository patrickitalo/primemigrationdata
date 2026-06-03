package prisma5

import (
	"github.com/primesoftwaresi/prime-migration/pkg/db/historydb"
)

// OptionStats contadores por opção de migração (UI / histórico).
type OptionStats struct {
	TotalOrigem int
	Skipped     int
	Novos       int
	Erros       int
}

// MigrationCallbacks integra migração PRISMA5 com o banco central de histórico.
type MigrationCallbacks struct {
	Incremental bool
	EntityType  int
	RunID       string
	ClientID    int
	RecordMeta  map[string]historydb.RecordMeta
	SaveRecord  func(runID string, clientID, entityType int, sourceID, destinationID, hash string) error
	SaveError   func(runID string, entityType int, sourceID, errMsg string) error
	Stats       *OptionStats
}
