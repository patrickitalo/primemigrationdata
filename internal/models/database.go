package models

type DatabaseConfig struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	Path      string `json:"path"`
	User      string `json:"user"`
	Password  string `json:"password"`
	Conversao string `json:"conversao"`
}
