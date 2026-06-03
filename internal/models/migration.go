package models

// MigrationMode define se a migração usa histórico central (incremental) ou não.
type MigrationMode string

const (
	MigrationModeCompleta    MigrationMode = "COMPLETA"
	MigrationModeIncremental MigrationMode = "INCREMENTAL"
)

type MigrationConfig struct {
	ClientCode        string         `json:"client_code"`
	System            string         `json:"system"`
	Database          DatabaseConfig `json:"database"`
	Options           []string       `json:"options"`
	Mode              MigrationMode  `json:"mode"`
	VVencido          *string        `json:"v_vencido"`
	ExcelPath         string         `json:"excel_path"`
	AliasPharmacie    string         `json:"alias_pharmacie"`
	IpServerPharmacie string         `json:"ipserver_pharmacie"`
	PortaPharmacie    string         `json:"porta_pharmacie"`
}

type MigrationStatus struct {
	Option     string `json:"option"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Duration   string `json:"duration"`
	Error      string `json:"error,omitempty"`
	TotalOrigem int   `json:"total_origem,omitempty"`
	Skipped     int   `json:"skipped,omitempty"`
	Novos       int   `json:"novos,omitempty"`
	ErrosReg    int   `json:"erros_registro,omitempty"`
}
