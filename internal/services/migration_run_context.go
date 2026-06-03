package services

// MigrationRunContext identifica o run no banco central durante uma opção de migração.
type MigrationRunContext struct {
	RunID       string
	ClientID    int
	Implantador string
}
